package domain

import "time"

// TransactionTemplate is a user-defined preset for one-tap transaction creation.
// Differs from RecurringTransaction: no schedule (frequency, next_run_at) — applied manually
// by the user from the dashboard, the AddTransaction page, or the Templates page.
type TransactionTemplate struct {
	ID            int64
	UserID        int64
	Name          string // empty -> UI generates from category + amount
	Type          TransactionType
	AmountCents   int64
	AmountFixed   bool // true: tap creates transaction immediately; false: prompt for amount
	CategoryID    int64
	AccountID     int64
	CurrencyCode  string
	Note          string
	SortOrder     int32
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Joined fields for display.
	CategoryName  string
	CategoryIcon  string
	CategoryColor string
}
