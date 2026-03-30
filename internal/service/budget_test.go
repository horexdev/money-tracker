package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newBudgetService(repo *mocks.MockBudgetStorer, txRepo *mocks.MockTransactionStorer) *service.BudgetService {
	return service.NewBudgetService(repo, txRepo, testutil.TestLogger())
}

func TestBudgetService_Create_ZeroLimit(t *testing.T) {
	svc := newBudgetService(&mocks.MockBudgetStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Create(context.Background(), &domain.Budget{LimitCents: 0, UserID: 1, CategoryID: 1})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestBudgetService_Create_DuplicateBudget(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})

	// GetByUserCategoryPeriod returns a budget (no error) → duplicate.
	existing := &domain.Budget{ID: 10}
	repo.On("GetByUserCategoryPeriod", mock.Anything, int64(1), int64(1), "monthly").Return(existing, nil)

	_, err := svc.Create(context.Background(), &domain.Budget{LimitCents: 10000, UserID: 1, CategoryID: 1, Period: "monthly"})
	assert.ErrorIs(t, err, domain.ErrBudgetAlreadyExists)
}

func TestBudgetService_Create_Success(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})

	// No existing budget.
	repo.On("GetByUserCategoryPeriod", mock.Anything, int64(1), int64(1), "monthly").Return(nil, domain.ErrBudgetNotFound)

	created := &domain.Budget{ID: 5, LimitCents: 10000}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Budget")).Return(created, nil)

	got, err := svc.Create(context.Background(), &domain.Budget{LimitCents: 10000, UserID: 1, CategoryID: 1, Period: "monthly"})
	require.NoError(t, err)
	assert.Equal(t, int64(5), got.ID)
}

func TestBudgetService_Update_ZeroLimit(t *testing.T) {
	svc := newBudgetService(&mocks.MockBudgetStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Update(context.Background(), &domain.Budget{LimitCents: 0})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestBudgetService_CheckAndNotify_NoNotifier(t *testing.T) {
	// WithNotifier never called → notifier is nil → should return nil without panicking.
	svc := newBudgetService(&mocks.MockBudgetStorer{}, &mocks.MockTransactionStorer{})
	err := svc.CheckAndNotify(context.Background(), 1)
	assert.NoError(t, err)
}

func TestBudgetService_CheckAndNotify_BelowThreshold(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})
	svc.WithNotifier(notifier)

	budget := &domain.Budget{
		ID:              1,
		UserID:          1,
		CategoryID:      1,
		LimitCents:      10000,
		SpentCents:      0, // 0% — below any threshold
		NotifyAtPercent: 80,
		Period:          domain.BudgetPeriodMonthly,
		CurrencyCode:    "USD",
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(0), nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertNotCalled(t, "SendBudgetAlert")
}

func TestBudgetService_CheckAndNotify_AlreadyNotifiedThisPeriod(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})
	svc.WithNotifier(notifier)

	// Budget is over threshold but was already notified this period.
	now := time.Now()
	budget := &domain.Budget{
		ID:              1,
		UserID:          1,
		CategoryID:      1,
		LimitCents:      10000,
		SpentCents:      9000, // 90% — above threshold
		NotifyAtPercent: 80,
		Period:          domain.BudgetPeriodMonthly,
		CurrencyCode:    "USD",
		LastNotifiedAt:  &now, // notified NOW → still in current period
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(9000), nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertNotCalled(t, "SendBudgetAlert")
}

func TestBudgetService_CheckAndNotify_SendsAlert(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})
	svc.WithNotifier(notifier)

	// Never notified before (LastNotifiedAt == nil).
	budget := &domain.Budget{
		ID:              1,
		UserID:          1,
		CategoryID:      1,
		CategoryName:    "Food",
		LimitCents:      10000,
		SpentCents:      9000, // 90% — above threshold
		NotifyAtPercent: 80,
		Period:          domain.BudgetPeriodMonthly,
		CurrencyCode:    "USD",
		LastNotifiedAt:  nil,
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(9000), nil)
	notifier.On("SendBudgetAlert", mock.Anything, int64(1), "Food", 90, int64(10000), int64(9000)).Return(nil)
	repo.On("UpdateLastNotified", mock.Anything, int64(1)).Return(nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertExpectations(t)
	repo.AssertCalled(t, "UpdateLastNotified", mock.Anything, int64(1))
}
