package service_test

import (
	"context"
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newBudgetService(repo *mocks.MockBudgetStorer, txRepo *mocks.MockTransactionStorer) *service.BudgetService {
	return service.NewBudgetService(repo, txRepo, &mocks.MockUserStorer{}, testutil.TestLogger())
}

func TestBudgetService_Create_ZeroLimit(t *testing.T) {
	svc := newBudgetService(&mocks.MockBudgetStorer{}, &mocks.MockTransactionStorer{})
	_, err := svc.Create(context.Background(), &domain.Budget{LimitCents: 0, UserID: 1, CategoryID: 1})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestBudgetService_Create_DuplicateBudget(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})

	existing := &domain.Budget{ID: 10}
	repo.On("GetByUserCategoryPeriod", mock.Anything, int64(1), int64(1), "monthly").Return(existing, nil)

	_, err := svc.Create(context.Background(), &domain.Budget{LimitCents: 10000, UserID: 1, CategoryID: 1, Period: "monthly"})
	assert.ErrorIs(t, err, domain.ErrBudgetAlreadyExists)
}

func TestBudgetService_Create_Success(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	svc := newBudgetService(repo, &mocks.MockTransactionStorer{})

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
	userRepo := &mocks.MockUserStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := service.NewBudgetService(repo, &mocks.MockTransactionStorer{}, userRepo, testutil.TestLogger())
	svc.WithNotifier(notifier)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "en"}, nil)

	budget := &domain.Budget{
		ID:                   1,
		UserID:               1,
		CategoryID:           1,
		LimitCents:           10000,
		SpentCents:           0, // 0% — below any threshold
		NotifyAtPercent:      80,
		NotificationsEnabled: true,
		Period:               domain.BudgetPeriodMonthly,
		CurrencyCode:         "USD",
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(0), nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertNotCalled(t, "SendBudgetAlert")
}

func TestBudgetService_CheckAndNotify_NotificationsDisabled(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	userRepo := &mocks.MockUserStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := service.NewBudgetService(repo, &mocks.MockTransactionStorer{}, userRepo, testutil.TestLogger())
	svc.WithNotifier(notifier)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "en"}, nil)

	budget := &domain.Budget{
		ID:                   1,
		UserID:               1,
		CategoryID:           1,
		LimitCents:           10000,
		SpentCents:           9000, // 90% — above threshold, but disabled
		NotifyAtPercent:      80,
		NotificationsEnabled: false,
		Period:               domain.BudgetPeriodMonthly,
		CurrencyCode:         "USD",
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(9000), nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertNotCalled(t, "SendBudgetAlert")
}

func TestBudgetService_CheckAndNotify_AlreadyNotifiedCurrentThreshold(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	userRepo := &mocks.MockUserStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := service.NewBudgetService(repo, &mocks.MockTransactionStorer{}, userRepo, testutil.TestLogger())
	svc.WithNotifier(notifier)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "ru"}, nil)

	// Already notified at 75% this period, usage is 80% — no new threshold crossed.
	budget := &domain.Budget{
		ID:                   1,
		UserID:               1,
		CategoryID:           1,
		LimitCents:           10000,
		SpentCents:           8000, // 80% → next fixed threshold is 95%, not yet reached
		NotifyAtPercent:      80,
		NotificationsEnabled: true,
		LastNotifiedPercent:  75,
		Period:               domain.BudgetPeriodMonthly,
		CurrencyCode:         "USD",
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(8000), nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertNotCalled(t, "SendBudgetAlert")
}

func TestBudgetService_CheckAndNotify_SendsAlert(t *testing.T) {
	repo := &mocks.MockBudgetStorer{}
	userRepo := &mocks.MockUserStorer{}
	notifier := &mocks.MockBudgetNotifier{}
	svc := service.NewBudgetService(repo, &mocks.MockTransactionStorer{}, userRepo, testutil.TestLogger())
	svc.WithNotifier(notifier)

	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "ru"}, nil)

	// Never notified (LastNotifiedPercent=0, LastNotifiedAt=nil), usage 55% → crosses 50% threshold.
	budget := &domain.Budget{
		ID:                   1,
		UserID:               1,
		CategoryID:           1,
		CategoryName:         "Food",
		LimitCents:           10000,
		SpentCents:           5500, // 55%
		NotifyAtPercent:      80,
		NotificationsEnabled: true,
		LastNotifiedPercent:  0,
		Period:               domain.BudgetPeriodMonthly,
		CurrencyCode:         "USD",
	}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{budget}, nil)
	repo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(1), "USD", mock.Anything, mock.Anything).Return(int64(5500), nil)
	notifier.On("SendBudgetAlert", mock.Anything, int64(1), "ru", "Food", "USD", 50, int64(10000), int64(5500)).Return(nil)
	repo.On("UpdateLastNotified", mock.Anything, int64(1), 50).Return(nil)

	err := svc.CheckAndNotify(context.Background(), 1)
	require.NoError(t, err)
	notifier.AssertExpectations(t)
	repo.AssertCalled(t, "UpdateLastNotified", mock.Anything, int64(1), 50)
}
