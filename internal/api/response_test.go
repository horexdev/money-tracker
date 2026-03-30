package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestHttpStatus(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{domain.ErrUserNotFound, http.StatusNotFound},
		{domain.ErrCategoryNotFound, http.StatusNotFound},
		{domain.ErrTransactionNotFound, http.StatusNotFound},
		{domain.ErrBudgetNotFound, http.StatusNotFound},
		{domain.ErrRecurringNotFound, http.StatusNotFound},
		{domain.ErrGoalNotFound, http.StatusNotFound},
		{domain.ErrAccountNotFound, http.StatusNotFound},
		{domain.ErrInvalidAmount, http.StatusBadRequest},
		{domain.ErrInvalidPeriod, http.StatusBadRequest},
		{domain.ErrInvalidCurrency, http.StatusBadRequest},
		{domain.ErrInvalidFrequency, http.StatusBadRequest},
		{domain.ErrInvalidLanguage, http.StatusBadRequest},
		{domain.ErrTooManyDisplayCurrencies, http.StatusBadRequest},
		{domain.ErrTransferSameAccount, http.StatusBadRequest},
		{domain.ErrBudgetAlreadyExists, http.StatusConflict},
		{domain.ErrCategoryInUse, http.StatusConflict},
		{domain.ErrAccountHasTransactions, http.StatusConflict},
		{domain.ErrCategorySystemReadOnly, http.StatusForbidden},
		{domain.ErrInsufficientGoalFunds, http.StatusUnprocessableEntity},
		{domain.ErrExchangeRateUnavailable, http.StatusServiceUnavailable},
		{errors.New("unexpected"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.status, api.HttpStatus(tt.err))
		})
	}
}

func TestUserMessage(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{domain.ErrUserNotFound, "user not found"},
		{domain.ErrCategoryNotFound, "category not found"},
		{domain.ErrTransactionNotFound, "transaction not found"},
		{domain.ErrBudgetNotFound, "budget not found"},
		{domain.ErrRecurringNotFound, "recurring transaction not found"},
		{domain.ErrGoalNotFound, "savings goal not found"},
		{domain.ErrAccountNotFound, "account not found"},
		{domain.ErrInvalidAmount, "invalid amount: must be a positive integer number of cents"},
		{domain.ErrInvalidCurrency, "invalid currency code"},
		{domain.ErrInvalidLanguage, "invalid language: use en or ru"},
		{domain.ErrBudgetAlreadyExists, "budget already exists for this category and period"},
		{domain.ErrCategoryInUse, "category has transactions and cannot be deleted"},
		{domain.ErrCategorySystemReadOnly, "system categories cannot be modified"},
		{domain.ErrInsufficientGoalFunds, "insufficient funds in savings goal"},
		{domain.ErrExchangeRateUnavailable, "exchange rate temporarily unavailable"},
		{domain.ErrAccountHasTransactions, "account has transactions and cannot be deleted"},
		{domain.ErrTransferSameAccount, "transfer source and destination must be different"},
		{errors.New("unknown"), "internal server error"},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			assert.Equal(t, tt.want, api.UserMessage(tt.err))
		})
	}
}
