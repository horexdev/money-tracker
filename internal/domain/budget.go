package domain

import "time"

// BudgetPeriod defines the time window for a budget limit.
type BudgetPeriod string

const (
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
)

// NotificationThresholds are the fixed percentages at which budget alerts are sent.
var NotificationThresholds = []int{50, 75, 95, 100}

// Budget represents a spending limit for a specific category within a time period.
type Budget struct {
	ID                   int64
	UserID               int64
	CategoryID           int64
	LimitCents           int64
	Period               BudgetPeriod
	CurrencyCode         string
	NotifyAtPercent      int
	NotificationsEnabled bool
	LastNotifiedPercent  int
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastNotifiedAt       *time.Time

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

// NextAlertThreshold returns the lowest fixed threshold that has been crossed
// but not yet notified, and 0 if none. Notifications must be enabled.
func (b *Budget) NextAlertThreshold() int {
	if !b.NotificationsEnabled {
		return 0
	}
	pct := int(b.UsagePercent())
	from, _ := b.periodBoundsNow()
	// If last notified was in a previous period, reset the notified percent.
	alreadyNotified := b.LastNotifiedPercent
	if b.LastNotifiedAt != nil && b.LastNotifiedAt.Before(from) {
		alreadyNotified = 0
	}
	for _, threshold := range NotificationThresholds {
		if pct >= threshold && alreadyNotified < threshold {
			return threshold
		}
	}
	return 0
}

// periodBoundsNow returns start of current period for this budget.
// Duplicate of service.periodBounds — kept here to avoid import cycle.
func (b *Budget) periodBoundsNow() (from, to time.Time) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch b.Period {
	case BudgetPeriodWeekly:
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		from = today.AddDate(0, 0, -(weekday - 1))
		to = from.AddDate(0, 0, 7)
	default:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 1, 0)
	}
	return from, to
}
