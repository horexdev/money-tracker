package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TestSentinelErrors_AreDistinct guards against accidental duplication of
// sentinel errors — two errors must never be `errors.Is`-equal.
func TestSentinelErrors_AreDistinct(t *testing.T) {
	all := []error{
		domain.ErrUserNotFound,
		domain.ErrCategoryNotFound,
		domain.ErrTransactionNotFound,
		domain.ErrBudgetNotFound,
		domain.ErrRecurringNotFound,
		domain.ErrGoalNotFound,
		domain.ErrInvalidAmount,
		domain.ErrInvalidPeriod,
		domain.ErrInvalidCurrency,
		domain.ErrInvalidFrequency,
		domain.ErrInvalidLanguage,
		domain.ErrTooManyDisplayCurrencies,
		domain.ErrExchangeRateUnavailable,
		domain.ErrCategoryNameEmpty,
		domain.ErrCategoryInUse,
		domain.ErrCategorySystemReadOnly,
		domain.ErrInsufficientGoalFunds,
		domain.ErrBudgetAlreadyExists,
		domain.ErrAccountNotFound,
		domain.ErrDefaultAccountExists,
		domain.ErrAccountHasTransactions,
		domain.ErrAccountHasTransfers,
		domain.ErrAccountHasRecurring,
		domain.ErrAccountHasTemplates,
		domain.ErrCategoryHasTemplates,
		domain.ErrTemplateNotFound,
		domain.ErrCannotDeleteLastAccount,
		domain.ErrMustSetNewDefault,
		domain.ErrCurrencyImmutable,
		domain.ErrTransferSameAccount,
		domain.ErrAdjustmentZeroAmount,
		domain.ErrCategoryProtected,
		domain.ErrInvalidSortParam,
	}
	for i, a := range all {
		for j, b := range all {
			if i == j {
				continue
			}
			assert.Falsef(t, errors.Is(a, b), "%v must not be Is-equal to %v", a, b)
		}
	}
}

// TestSentinelErrors_HaveNonEmptyMessages ensures every public error carries
// a human-readable message — empty messages cause confusing API responses.
func TestSentinelErrors_HaveNonEmptyMessages(t *testing.T) {
	all := []error{
		domain.ErrUserNotFound,
		domain.ErrCategoryNotFound,
		domain.ErrTransactionNotFound,
		domain.ErrBudgetNotFound,
		domain.ErrRecurringNotFound,
		domain.ErrGoalNotFound,
		domain.ErrInvalidAmount,
		domain.ErrInvalidPeriod,
		domain.ErrInvalidCurrency,
		domain.ErrInvalidFrequency,
		domain.ErrInvalidLanguage,
		domain.ErrTooManyDisplayCurrencies,
		domain.ErrExchangeRateUnavailable,
		domain.ErrCategoryNameEmpty,
		domain.ErrCategoryInUse,
		domain.ErrCategorySystemReadOnly,
		domain.ErrInsufficientGoalFunds,
		domain.ErrBudgetAlreadyExists,
		domain.ErrAccountNotFound,
		domain.ErrDefaultAccountExists,
		domain.ErrAccountHasTransactions,
		domain.ErrAccountHasTransfers,
		domain.ErrAccountHasRecurring,
		domain.ErrAccountHasTemplates,
		domain.ErrCategoryHasTemplates,
		domain.ErrTemplateNotFound,
		domain.ErrCannotDeleteLastAccount,
		domain.ErrMustSetNewDefault,
		domain.ErrCurrencyImmutable,
		domain.ErrTransferSameAccount,
		domain.ErrAdjustmentZeroAmount,
		domain.ErrCategoryProtected,
		domain.ErrInvalidSortParam,
	}
	for _, e := range all {
		assert.NotEmpty(t, e.Error(), "error must have a non-empty message")
	}
}
