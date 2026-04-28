package scheduler

// Tick exposes tick for use in tests without changing the production API.
func (s *Scheduler) Tick() {
	s.tick(nil)
}

// SaveSnapshots exposes saveSnapshots for tests.
func (s *Scheduler) SaveSnapshots() {
	s.saveSnapshots(nil)
}

// CheckBudgets exposes checkBudgets for tests.
func (s *Scheduler) CheckBudgets() {
	s.checkBudgets(nil)
}

// LastSnapshot returns the cached last successful snapshot date.
func (s *Scheduler) LastSnapshot() any { return s.lastSnapshot }
