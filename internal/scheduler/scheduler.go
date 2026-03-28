package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// RecurringProcessor creates transactions for due recurring items.
type RecurringProcessor interface {
	ProcessDue(ctx context.Context) (int, error)
}

// Scheduler runs periodic background jobs.
type Scheduler struct {
	recurring RecurringProcessor
	log       *slog.Logger
	interval  time.Duration
}

// New creates a Scheduler that ticks every interval.
func New(recurring RecurringProcessor, log *slog.Logger, interval time.Duration) *Scheduler {
	return &Scheduler{
		recurring: recurring,
		log:       log,
		interval:  interval,
	}
}

// Run blocks until ctx is cancelled, processing recurring transactions each tick.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	s.log.Info("scheduler started", slog.Duration("interval", s.interval))

	for {
		select {
		case <-ctx.Done():
			s.log.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	processed, err := s.recurring.ProcessDue(ctx)
	if err != nil {
		s.log.Error("scheduler: process recurring failed", slog.String("error", err.Error()))
		return
	}
	if processed > 0 {
		s.log.Info("scheduler: recurring transactions processed", slog.Int("count", processed))
	}
}
