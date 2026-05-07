package mocks

import (
	"context"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/mock"
)

// MockTransactionAdder is a testify mock for service.TransactionAdder.
type MockTransactionAdder struct {
	mock.Mock
}

func (m *MockTransactionAdder) AddExpense(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode string, accountID int64, createdAt *time.Time) (*domain.Transaction, error) {
	args := m.Called(ctx, userID, amountCents, categoryID, note, currencyCode, accountID, createdAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}

func (m *MockTransactionAdder) AddIncome(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode string, accountID int64, createdAt *time.Time) (*domain.Transaction, error) {
	args := m.Called(ctx, userID, amountCents, categoryID, note, currencyCode, accountID, createdAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Transaction), args.Error(1)
}
