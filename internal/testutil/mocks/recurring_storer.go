package mocks

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockRecurringStorer is a testify mock for service.RecurringStorer.
type MockRecurringStorer struct {
	mock.Mock
}

func (m *MockRecurringStorer) Create(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	args := m.Called(ctx, rt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) GetByID(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) ListByUser(ctx context.Context, userID int64) ([]*domain.RecurringTransaction, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) Update(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	args := m.Called(ctx, rt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) ToggleActive(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRecurringStorer) GetDue(ctx context.Context, before time.Time) ([]*domain.RecurringTransaction, error) {
	args := m.Called(ctx, before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RecurringTransaction), args.Error(1)
}

func (m *MockRecurringStorer) UpdateNextRun(ctx context.Context, id int64, nextRun time.Time) error {
	args := m.Called(ctx, id, nextRun)
	return args.Error(0)
}
