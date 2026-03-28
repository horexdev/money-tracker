package domain

import "time"

// SavingsGoal represents a financial target the user is saving towards.
type SavingsGoal struct {
	ID           int64
	UserID       int64
	Name         string
	TargetCents  int64
	CurrentCents int64
	CurrencyCode string
	Deadline     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ProgressPercent returns the completion percentage of the goal.
func (g *SavingsGoal) ProgressPercent() float64 {
	if g.TargetCents == 0 {
		return 0
	}
	p := float64(g.CurrentCents) / float64(g.TargetCents) * 100
	if p > 100 {
		return 100
	}
	return p
}

// IsCompleted returns true if the goal has been reached.
func (g *SavingsGoal) IsCompleted() bool {
	return g.CurrentCents >= g.TargetCents
}

// RemainingCents returns how much more is needed to reach the target.
func (g *SavingsGoal) RemainingCents() int64 {
	remaining := g.TargetCents - g.CurrentCents
	if remaining < 0 {
		return 0
	}
	return remaining
}
