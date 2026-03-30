package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockSavingsGoalStorer is a testify mock for service.SavingsGoalStorer.
type MockSavingsGoalStorer struct {
	mock.Mock
}

func (m *MockSavingsGoalStorer) Create(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, g)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) GetByID(ctx context.Context, id, userID int64) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) ListByUser(ctx context.Context, userID int64) ([]*domain.SavingsGoal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) Update(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, g)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, id, userID, amountCents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	args := m.Called(ctx, id, userID, amountCents)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SavingsGoal), args.Error(1)
}

func (m *MockSavingsGoalStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockSavingsGoalStorer) ListHistory(ctx context.Context, goalID, userID int64) ([]*domain.GoalTransaction, error) {
	args := m.Called(ctx, goalID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GoalTransaction), args.Error(1)
}

func (m *MockSavingsGoalStorer) GetByAccountID(ctx context.Context, accountID int64) ([]*domain.SavingsGoal, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.SavingsGoal), args.Error(1)
}
