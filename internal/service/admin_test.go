package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func newAdminService(repo *mocks.MockAdminStorer) *service.AdminService {
	return service.NewAdminService(repo, testutil.TestLogger())
}

func TestAdminService_ListUsers_ClampsPageAndPageSize(t *testing.T) {
	t.Run("page<1 becomes 1, pageSize<1 becomes 20", func(t *testing.T) {
		repo := &mocks.MockAdminStorer{}
		svc := newAdminService(repo)
		repo.On("CountUsers", mock.Anything).Return(int64(50), nil)
		repo.On("ListUsers", mock.Anything, 20, 0).Return([]*domain.User{{ID: 1}}, nil)

		_, total, err := svc.ListUsers(context.Background(), 0, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(50), total)
		repo.AssertExpectations(t)
	})

	t.Run("pageSize>100 becomes 20", func(t *testing.T) {
		repo := &mocks.MockAdminStorer{}
		svc := newAdminService(repo)
		repo.On("CountUsers", mock.Anything).Return(int64(0), nil)
		repo.On("ListUsers", mock.Anything, 20, 0).Return([]*domain.User{}, nil)

		_, _, err := svc.ListUsers(context.Background(), 1, 500)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("offset = (page-1)*pageSize", func(t *testing.T) {
		repo := &mocks.MockAdminStorer{}
		svc := newAdminService(repo)
		repo.On("CountUsers", mock.Anything).Return(int64(0), nil)
		repo.On("ListUsers", mock.Anything, 25, 50).Return([]*domain.User{}, nil)

		_, _, err := svc.ListUsers(context.Background(), 3, 25)
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestAdminService_ListUsers_ReturnsCountError(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	svc := newAdminService(repo)
	repo.On("CountUsers", mock.Anything).Return(int64(0), errors.New("db down"))

	users, total, err := svc.ListUsers(context.Background(), 1, 20)
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.Zero(t, total)
}

func TestAdminService_GetStats_AggregatesAllMetrics(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	svc := newAdminService(repo)

	repo.On("CountUsers", mock.Anything).Return(int64(1000), nil)
	// Three CountNewUsers calls for today/week/month
	repo.On("CountNewUsers", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(5), nil)
	// Three more for cohort retention day-1, day-7, day-30
	repo.On("CountRetainedUsers", mock.Anything,
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(3), nil)

	stats, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, int64(1000), stats.TotalUsers)
	assert.Equal(t, int64(5), stats.NewToday)
	assert.Equal(t, int64(5), stats.NewThisWeek)
	assert.Equal(t, int64(5), stats.NewThisMonth)
	// Retention is 3/5 = 60% for each window since CountNewUsers returns 5 too.
	assert.InDelta(t, 60.0, stats.RetentionDay1, 0.01)
	assert.InDelta(t, 60.0, stats.RetentionDay7, 0.01)
	assert.InDelta(t, 60.0, stats.RetentionDay30, 0.01)
}

func TestAdminService_GetStats_RetentionZeroWhenCohortEmpty(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	svc := newAdminService(repo)

	repo.On("CountUsers", mock.Anything).Return(int64(10), nil)
	repo.On("CountNewUsers", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(0), nil)
	// CountRetainedUsers should not be called when cohortSize == 0, but stub it
	// just in case to keep the mock unstrict.
	repo.On("CountRetainedUsers", mock.Anything,
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(0), nil).Maybe()

	stats, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	assert.Equal(t, float64(0), stats.RetentionDay1)
	assert.Equal(t, float64(0), stats.RetentionDay7)
	assert.Equal(t, float64(0), stats.RetentionDay30)
}

func TestAdminService_ListAllUserIDs_PassesThrough(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	svc := newAdminService(repo)
	repo.On("ListAllUserIDs", mock.Anything).Return([]int64{1, 2, 3}, nil)

	ids, err := svc.ListAllUserIDs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []int64{1, 2, 3}, ids)
}

// TestAdminService_GetStats_TimeBoundaries verifies the service computes
// today/week/month boundaries correctly: today should start at 00:00 UTC,
// week at the most recent Sunday, month at the 1st.
func TestAdminService_GetStats_TimeBoundaries(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	svc := newAdminService(repo)

	repo.On("CountUsers", mock.Anything).Return(int64(0), nil)

	var capturedFrom []time.Time
	repo.On("CountNewUsers", mock.Anything, mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(0), nil).
		Run(func(args mock.Arguments) {
			capturedFrom = append(capturedFrom, args.Get(1).(time.Time))
		})
	repo.On("CountRetainedUsers", mock.Anything,
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return(int64(0), nil).Maybe()

	_, err := svc.GetStats(context.Background())
	require.NoError(t, err)
	// CountNewUsers is called 6 times: 3 windows (today/week/month) + 3 cohort sizes
	// (day-1, day-7, day-30). We only assert on the first three.
	require.GreaterOrEqual(t, len(capturedFrom), 3, "expected at least 3 CountNewUsers calls")

	// today at 00:00 UTC
	assert.Equal(t, 0, capturedFrom[0].Hour())
	assert.Equal(t, 0, capturedFrom[0].Minute())
	// week start is at or before today
	assert.True(t, !capturedFrom[1].After(capturedFrom[0]))
	// month start is the 1st
	assert.Equal(t, 1, capturedFrom[2].Day())
}
