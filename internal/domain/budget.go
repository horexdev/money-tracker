package domain

import "time"

// BudgetPeriod defines the time window for a budget limit.
type BudgetPeriod string

const (
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

// Budget represents a spending limit for a specific category within a time period.
type Budget struct {
	ID              int64
	UserID          int64
	CategoryID      int64
	LimitCents      int64
	Period          BudgetPeriod
	CurrencyCode    string
	NotifyAtPercent int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastNotifiedAt  *time.Time

	// Joined fields for display.
	CategoryName  string
	CategoryEmoji string
	CategoryColor string
	SpentCents    int64
}

// UsagePercent returns the percentage of the budget that has been spent.
func (b *Budget) UsagePercent() float64 {
	if b.LimitCents == 0 {
		return 0
	}
	return float64(b.SpentCents) / float64(b.LimitCents) * 100
}

// IsOverLimit returns true if spending has exceeded the budget limit.
func (b *Budget) IsOverLimit() bool {
	return b.SpentCents >= b.LimitCents
}

// ShouldNotify returns true if spending has reached the notification threshold.
func (b *Budget) ShouldNotify() bool {
	return b.UsagePercent() >= float64(b.NotifyAtPercent)
}
