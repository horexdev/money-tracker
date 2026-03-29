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

// BudgetChecker sends budget threshold notifications for all users with budgets.
type BudgetChecker interface {
	ListDistinctUserIDs(ctx context.Context) ([]int64, error)
	CheckAndNotify(ctx context.Context, userID int64) error
}

// Scheduler runs periodic background jobs.
type Scheduler struct {
	recurring RecurringProcessor
	budgets   BudgetChecker
	log       *slog.Logger
	interval  time.Duration
}

// New creates a Scheduler that ticks every interval.
func New(recurring RecurringProcessor, budgets BudgetChecker, log *slog.Logger, interval time.Duration) *Scheduler {
	return &Scheduler{
		recurring: recurring,
		budgets:   budgets,
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
	} else if processed > 0 {
		s.log.Info("scheduler: recurring transactions processed", slog.Int("count", processed))
	}

	if s.budgets != nil {
		s.checkBudgets(ctx)
	}
}

func (s *Scheduler) checkBudgets(ctx context.Context) {
	userIDs, err := s.budgets.ListDistinctUserIDs(ctx)
	if err != nil {
		s.log.Error("scheduler: list budget users failed", slog.String("error", err.Error()))
		return
	}

	for _, uid := range userIDs {
		if err := s.budgets.CheckAndNotify(ctx, uid); err != nil {
			s.log.Warn("scheduler: budget check failed",
				slog.Int64("user_id", uid),
				slog.String("error", err.Error()),
			)
		}
	}
}
