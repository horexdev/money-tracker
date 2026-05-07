package scheduler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/scheduler"
	"github.com/horexdev/money-tracker/internal/testutil"
)

type fakeRecurring struct{ mock.Mock }

func (f *fakeRecurring) ProcessDue(ctx context.Context) (int, error) {
	args := f.Called(ctx)
	return args.Int(0), args.Error(1)
}

type fakeBudgets struct{ mock.Mock }

func (f *fakeBudgets) ListDistinctUserIDs(ctx context.Context) ([]int64, error) {
	args := f.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

func (f *fakeBudgets) CheckAndNotify(ctx context.Context, userID int64) error {
	args := f.Called(ctx, userID)
	return args.Error(0)
}

type fakeSnapshots struct{ mock.Mock }

func (f *fakeSnapshots) SaveDaily(ctx context.Context) error {
	args := f.Called(ctx)
	return args.Error(0)
}

func TestScheduler_Tick_RunsAllJobs(t *testing.T) {
	rec := &fakeRecurring{}
	bud := &fakeBudgets{}
	snap := &fakeSnapshots{}

	rec.On("ProcessDue", mock.Anything).Return(2, nil)
	bud.On("ListDistinctUserIDs", mock.Anything).Return([]int64{10, 20}, nil)
	bud.On("CheckAndNotify", mock.Anything, int64(10)).Return(nil)
	bud.On("CheckAndNotify", mock.Anything, int64(20)).Return(nil)
	snap.On("SaveDaily", mock.Anything).Return(nil)

	s := scheduler.New(rec, bud, snap, testutil.TestLogger(), time.Hour)
	s.Tick()

	rec.AssertExpectations(t)
	bud.AssertExpectations(t)
	snap.AssertCalled(t, "SaveDaily", mock.Anything)
}

func TestScheduler_Tick_ContinuesOnRecurringError(t *testing.T) {
	rec := &fakeRecurring{}
	snap := &fakeSnapshots{}
	rec.On("ProcessDue", mock.Anything).Return(0, errors.New("db down"))
	snap.On("SaveDaily", mock.Anything).Return(nil)

	s := scheduler.New(rec, nil, snap, testutil.TestLogger(), time.Hour)
	s.Tick() // must not panic, must still call snapshots
	snap.AssertCalled(t, "SaveDaily", mock.Anything)
}

func TestScheduler_Tick_NilBudgetsSkipped(t *testing.T) {
	rec := &fakeRecurring{}
	snap := &fakeSnapshots{}
	rec.On("ProcessDue", mock.Anything).Return(0, nil)
	snap.On("SaveDaily", mock.Anything).Return(nil)

	s := scheduler.New(rec, nil, snap, testutil.TestLogger(), time.Hour)
	s.Tick()
	rec.AssertExpectations(t)
}

func TestScheduler_SaveSnapshots_Idempotent(t *testing.T) {
	snap := &fakeSnapshots{}
	snap.On("SaveDaily", mock.Anything).Return(nil).Once()

	s := scheduler.New(&fakeRecurring{}, nil, snap, testutil.TestLogger(), time.Hour)
	s.SaveSnapshots() // first call: invokes SaveDaily
	s.SaveSnapshots() // second call same day: must not invoke again
	snap.AssertNumberOfCalls(t, "SaveDaily", 1)
}

func TestScheduler_CheckBudgets_ListErrorIsLoggedAndIgnored(t *testing.T) {
	bud := &fakeBudgets{}
	bud.On("ListDistinctUserIDs", mock.Anything).Return(nil, errors.New("db"))

	s := scheduler.New(&fakeRecurring{}, bud, nil, testutil.TestLogger(), time.Hour)
	s.CheckBudgets()
	bud.AssertNotCalled(t, "CheckAndNotify", mock.Anything, mock.Anything)
}

func TestScheduler_CheckBudgets_PerUserErrorContinues(t *testing.T) {
	bud := &fakeBudgets{}
	bud.On("ListDistinctUserIDs", mock.Anything).Return([]int64{1, 2, 3}, nil)
	bud.On("CheckAndNotify", mock.Anything, int64(1)).Return(nil)
	bud.On("CheckAndNotify", mock.Anything, int64(2)).Return(errors.New("user 2 failed"))
	bud.On("CheckAndNotify", mock.Anything, int64(3)).Return(nil)

	s := scheduler.New(&fakeRecurring{}, bud, nil, testutil.TestLogger(), time.Hour)
	s.CheckBudgets()
	bud.AssertNumberOfCalls(t, "CheckAndNotify", 3)
}

func TestScheduler_Run_StopsOnContextCancel(t *testing.T) {
	rec := &fakeRecurring{}
	rec.On("ProcessDue", mock.Anything).Return(0, nil)
	snap := &fakeSnapshots{}
	snap.On("SaveDaily", mock.Anything).Return(nil)

	s := scheduler.New(rec, nil, snap, testutil.TestLogger(), 50*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()
	time.Sleep(120 * time.Millisecond)
	cancel()
	select {
	case <-done:
		// ok — Run returned promptly after cancel
	case <-time.After(time.Second):
		require.Fail(t, "Run did not return within 1s of cancel")
	}
	assert.GreaterOrEqual(t, len(rec.Calls), 1, "Run should have ticked at least once")
}
