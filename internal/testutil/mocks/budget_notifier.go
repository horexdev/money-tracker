package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockBudgetNotifier is a testify mock for service.BudgetNotifier.
type MockBudgetNotifier struct {
	mock.Mock
}

func (m *MockBudgetNotifier) SendBudgetAlert(ctx context.Context, chatID int64, lang, categoryName, currencyCode string, spentPercent int, limitCents, spentCents int64) error {
	args := m.Called(ctx, chatID, lang, categoryName, currencyCode, spentPercent, limitCents, spentCents)
	return args.Error(0)
}
