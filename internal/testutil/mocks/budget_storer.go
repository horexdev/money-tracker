package mocks

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockBudgetStorer is a testify mock for service.BudgetStorer.
type MockBudgetStorer struct {
	mock.Mock
}

func (m *MockBudgetStorer) Create(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	args := m.Called(ctx, b)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetStorer) GetByID(ctx context.Context, id, userID int64) (*domain.Budget, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetStorer) ListByUser(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Budget), args.Error(1)
}

func (m *MockBudgetStorer) Update(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	args := m.Called(ctx, b)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockBudgetStorer) GetByUserCategoryPeriod(ctx context.Context, userID, categoryID int64, period string) (*domain.Budget, error) {
	args := m.Called(ctx, userID, categoryID, period)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Budget), args.Error(1)
}

func (m *MockBudgetStorer) GetSpentInPeriod(ctx context.Context, userID, categoryID int64, currency string, from, to time.Time) (int64, error) {
	args := m.Called(ctx, userID, categoryID, currency, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockBudgetStorer) UpdateLastNotified(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockBudgetStorer) ListDistinctUserIDs(ctx context.Context) ([]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}
