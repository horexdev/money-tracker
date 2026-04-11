package mocks

import (
	"context"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockAccountStorer is a testify mock for service.AccountStorer.
type MockAccountStorer struct {
	mock.Mock
}

func (m *MockAccountStorer) Create(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	args := m.Called(ctx, a)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) GetByID(ctx context.Context, id, userID int64) (*domain.Account, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) GetDefault(ctx context.Context, userID int64) (*domain.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) ListByUser(ctx context.Context, userID int64) ([]*domain.Account, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) Update(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	args := m.Called(ctx, a)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) SetDefault(ctx context.Context, id, userID int64) (*domain.Account, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Account), args.Error(1)
}

func (m *MockAccountStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockAccountStorer) CountTransactions(ctx context.Context, accountID, userID int64) (int64, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAccountStorer) GetBalance(ctx context.Context, accountID, userID int64) (int64, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAccountStorer) CountAccounts(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAccountStorer) CountTransfers(ctx context.Context, accountID, userID int64) (int64, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAccountStorer) CountRecurring(ctx context.Context, accountID, userID int64) (int64, error) {
	args := m.Called(ctx, accountID, userID)
	return args.Get(0).(int64), args.Error(1)
}
