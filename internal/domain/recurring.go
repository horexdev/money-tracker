package domain

import "time"

// Frequency defines how often a recurring transaction repeats.
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

// RecurringTransaction is a template for automatically repeated transactions.
type RecurringTransaction struct {
	ID            int64
	UserID        int64
	CategoryID    int64
	Type          TransactionType
	AmountCents   int64
	CurrencyCode  string
	Note          string
	Frequency     Frequency
	NextRunAt     time.Time
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Joined fields for display.
	CategoryName  string
	CategoryEmoji string
	CategoryColor string
}

// NextRunAfter calculates the next run time after the given time based on frequency.
func (r *RecurringTransaction) NextRunAfter(after time.Time) time.Time {
	switch r.Frequency {
	case FrequencyDaily:
		return after.AddDate(0, 0, 1)
	case FrequencyWeekly:
		return after.AddDate(0, 0, 7)
	case FrequencyMonthly:
		return after.AddDate(0, 1, 0)
	case FrequencyYearly:
		return after.AddDate(1, 0, 0)
	default:
		return after.AddDate(0, 1, 0)
	}
}
