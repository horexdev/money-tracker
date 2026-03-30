package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockTransferStorer is a testify mock for service.TransferStorer.
type MockTransferStorer struct {
	mock.Mock
}

func (m *MockTransferStorer) Create(ctx context.Context, t *domain.Transfer) (*domain.Transfer, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transfer), args.Error(1)
}

func (m *MockTransferStorer) GetByID(ctx context.Context, id, userID int64) (*domain.Transfer, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transfer), args.Error(1)
}

func (m *MockTransferStorer) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transfer, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transfer), args.Error(1)
}

func (m *MockTransferStorer) ListByAccount(ctx context.Context, userID, accountID int64) ([]*domain.Transfer, error) {
	args := m.Called(ctx, userID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transfer), args.Error(1)
}

func (m *MockTransferStorer) Count(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransferStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}
