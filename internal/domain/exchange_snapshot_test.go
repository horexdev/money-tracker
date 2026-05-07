package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/domain"
)

func TestExchangeRateSnapshot_ZeroValue(t *testing.T) {
	var snap domain.ExchangeRateSnapshot
	assert.Zero(t, snap.ID)
	assert.True(t, snap.SnapshotDate.IsZero())
	assert.Empty(t, snap.BaseCurrency)
	assert.Empty(t, snap.TargetCurrency)
	assert.Zero(t, snap.Rate)
}

func TestExchangeRateSnapshot_AssignAllFields(t *testing.T) {
	snap := domain.ExchangeRateSnapshot{
		ID:             1,
		SnapshotDate:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		BaseCurrency:   "USD",
		TargetCurrency: "EUR",
		Rate:           0.92,
		CreatedAt:      time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}
	assert.Equal(t, int64(1), snap.ID)
	assert.Equal(t, "USD", snap.BaseCurrency)
	assert.Equal(t, "EUR", snap.TargetCurrency)
	assert.InDelta(t, 0.92, snap.Rate, 0.0001)
}
