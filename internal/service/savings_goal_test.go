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

func newGoalService(repo *mocks.MockSavingsGoalStorer) *service.SavingsGoalService {
	return service.NewSavingsGoalService(repo, testutil.TestLogger())
}

func TestSavingsGoalService_Create_ZeroTarget(t *testing.T) {
	svc := newGoalService(&mocks.MockSavingsGoalStorer{})
	_, err := svc.Create(context.Background(), &domain.SavingsGoal{TargetCents: 0})
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestSavingsGoalService_Create_Success(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	svc := newGoalService(repo)

	g := &domain.SavingsGoal{TargetCents: 50000, Name: "Vacation"}
	created := &domain.SavingsGoal{ID: 1, TargetCents: 50000, Name: "Vacation"}
	repo.On("Create", mock.Anything, g).Return(created, nil)

	got, err := svc.Create(context.Background(), g)
	require.NoError(t, err)
	assert.Equal(t, int64(1), got.ID)
}

func TestSavingsGoalService_Deposit_ZeroAmount(t *testing.T) {
	svc := newGoalService(&mocks.MockSavingsGoalStorer{})
	_, err := svc.Deposit(context.Background(), 1, 1, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestSavingsGoalService_Deposit_ValidAmount(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	svc := newGoalService(repo)

	updated := &domain.SavingsGoal{ID: 1, CurrentCents: 5000}
	repo.On("Deposit", mock.Anything, int64(1), int64(42), int64(5000)).Return(updated, nil)

	got, err := svc.Deposit(context.Background(), 1, 42, 5000)
	require.NoError(t, err)
	assert.Equal(t, int64(5000), got.CurrentCents)
}

func TestSavingsGoalService_Withdraw_ZeroAmount(t *testing.T) {
	svc := newGoalService(&mocks.MockSavingsGoalStorer{})
	_, err := svc.Withdraw(context.Background(), 1, 1, 0)
	assert.ErrorIs(t, err, domain.ErrInvalidAmount)
}

func TestSavingsGoalService_Withdraw_InsufficientFunds(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	svc := newGoalService(repo)

	repo.On("Withdraw", mock.Anything, int64(1), int64(42), int64(99999)).Return(nil, domain.ErrInsufficientGoalFunds)

	_, err := svc.Withdraw(context.Background(), 1, 42, 99999)
	assert.Error(t, err)
}
