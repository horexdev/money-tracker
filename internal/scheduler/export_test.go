package scheduler

import "context"

// Tick exposes tick for use in tests without changing the production API.
func (s *Scheduler) Tick() {
	s.tick(context.Background())
}

// SaveSnapshots exposes saveSnapshots for tests.
func (s *Scheduler) SaveSnapshots() {
	s.saveSnapshots(context.Background())
}

// CheckBudgets exposes checkBudgets for tests.
func (s *Scheduler) CheckBudgets() {
	s.checkBudgets(context.Background())
}

// LastSnapshot returns the cached last successful snapshot date.
func (s *Scheduler) LastSnapshot() any { return s.lastSnapshot }
