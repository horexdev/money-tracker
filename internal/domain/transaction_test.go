package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TestTransactionType_StringValues guards the TransactionType constants against
// accidental edits — the database transaction_type enum has 'expense'/'income'.
func TestTransactionType_StringValues(t *testing.T) {
	assert.Equal(t, "expense", string(domain.TransactionTypeExpense))
	assert.Equal(t, "income", string(domain.TransactionTypeIncome))
}

func TestTransaction_DefaultValues(t *testing.T) {
	tx := &domain.Transaction{}
	assert.Zero(t, tx.ID)
	assert.Zero(t, tx.AmountCents)
	assert.False(t, tx.IsAdjustment, "adjustments must be opt-in via the IsAdjustment flag")
	assert.Empty(t, string(tx.Type), "Type defaults to empty until explicitly set")
}
