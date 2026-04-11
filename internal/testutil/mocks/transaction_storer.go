package mocks

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockTransactionStorer is a testify mock for service.TransactionStorer.
type MockTransactionStorer struct {
	mock.Mock
}

func (m *MockTransactionStorer) Create(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) CreateWithDate(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) Delete(ctx context.Context, id, userID int64) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTransactionStorer) GetBalance(ctx context.Context, userID int64) (int64, int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Get(1).(int64), args.Error(2)
}

func (m *MockTransactionStorer) GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BalanceByCurrency), args.Error(1)
}

func (m *MockTransactionStorer) GetTotalInBaseCurrency(ctx context.Context, userID int64, targetCurrency string) (int64, error) {
	args := m.Called(ctx, userID, targetCurrency)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionStorer) List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) Count(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionStorer) StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	args := m.Called(ctx, userID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CategoryStat), args.Error(1)
}

func (m *MockTransactionStorer) ListByCategoryPeriod(ctx context.Context, userID, categoryID int64, from, to time.Time) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, categoryID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) Update(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) ListByAccount(ctx context.Context, userID, accountID int64, limit, offset int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, accountID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) CountByAccount(ctx context.Context, userID, accountID int64) (int64, error) {
	args := m.Called(ctx, userID, accountID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionStorer) StatsByCategoryAndAccount(ctx context.Context, userID, accountID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	args := m.Called(ctx, userID, accountID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CategoryStat), args.Error(1)
}

func (m *MockTransactionStorer) GetBalanceByCurrencyAndAccount(ctx context.Context, userID, accountID int64) ([]domain.BalanceByCurrency, error) {
	args := m.Called(ctx, userID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.BalanceByCurrency), args.Error(1)
}

func (m *MockTransactionStorer) CreateAdjustment(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) ListWithDateRange(ctx context.Context, userID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, from, to, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) ListByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error) {
	args := m.Called(ctx, userID, accountID, from, to, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Transaction), args.Error(1)
}

func (m *MockTransactionStorer) CountWithDateRange(ctx context.Context, userID int64, from, to *time.Time) (int64, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTransactionStorer) CountByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time) (int64, error) {
	args := m.Called(ctx, userID, accountID, from, to)
	return args.Get(0).(int64), args.Error(1)
}
