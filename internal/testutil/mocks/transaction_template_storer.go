package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockTransactionTemplateStorer is a testify mock for service.TransactionTemplateStorer.
type MockTransactionTemplateStorer struct {
	mock.Mock
}

func (m *MockTransactionTemplateStorer) Create(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TransactionTemplate), args.Error(1)
}

func (m *MockTransactionTemplateStorer) GetByID(ctx context.Context, id, userID int64) (*domain.TransactionTemplate, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TransactionTemplate), args.Error(1)
}

func (m *MockTransactionTemplateStorer) ListByUser(ctx context.Context, userID int64) ([]*domain.TransactionTemplate, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TransactionTemplate), args.Error(1)
}

func (m *MockTransactionTemplateStorer) Update(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TransactionTemplate), args.Error(1)
}

func (m *MockTransactionTemplateStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTransactionTemplateStorer) Reorder(ctx context.Context, userID int64, orderedIDs []int64) error {
	args := m.Called(ctx, userID, orderedIDs)
	return args.Error(0)
}
