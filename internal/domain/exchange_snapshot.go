package domain

import "time"

// ExchangeRateSnapshot is a historical exchange rate for a currency pair on a given date.
type ExchangeRateSnapshot struct {
	ID             int64
	SnapshotDate   time.Time
	BaseCurrency   string
	TargetCurrency string
	Rate           float64
	CreatedAt      time.Time
}
