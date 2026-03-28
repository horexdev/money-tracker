package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// BudgetService handles business logic for budgets.
type BudgetService struct {
	repo BudgetStorer
	log  *slog.Logger
}

func NewBudgetService(repo BudgetStorer, log *slog.Logger) *BudgetService {
	return &BudgetService{repo: repo, log: log}
}

// Create adds a new budget for a category.
func (s *BudgetService) Create(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	if b.LimitCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Check for duplicate.
	_, err := s.repo.GetByUserCategoryPeriod(ctx, b.UserID, b.CategoryID, string(b.Period))
	if err == nil {
		return nil, domain.ErrBudgetAlreadyExists
	}
	if !errors.Is(err, domain.ErrBudgetNotFound) {
		return nil, fmt.Errorf("check existing budget: %w", err)
	}

	budget, err := s.repo.Create(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}

	s.log.InfoContext(ctx, "budget created",
		slog.Int64("user_id", b.UserID),
		slog.Int64("budget_id", budget.ID),
		slog.Int64("category_id", b.CategoryID),
	)
	return budget, nil
}

// GetByID returns a budget scoped to user.
func (s *BudgetService) GetByID(ctx context.Context, id, userID int64) (*domain.Budget, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListWithProgress returns all budgets for a user with spent amounts populated.
func (s *BudgetService) ListWithProgress(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	budgets, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list budgets for user %d: %w", userID, err)
	}

	now := time.Now()
	for _, b := range budgets {
		from, to := periodBounds(b.Period, now)
		spent, err := s.repo.GetSpentInPeriod(ctx, b.UserID, b.CategoryID, b.CurrencyCode, from, to)
		if err != nil {
			s.log.WarnContext(ctx, "failed to get spent for budget",
				slog.Int64("budget_id", b.ID),
				slog.String("error", err.Error()),
			)
			continue
		}
		b.SpentCents = spent
	}
	return budgets, nil
}

// Update modifies an existing budget.
func (s *BudgetService) Update(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	if b.LimitCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	budget, err := s.repo.Update(ctx, b)
	if err != nil {
		return nil, fmt.Errorf("update budget %d: %w", b.ID, err)
	}
	return budget, nil
}

// Delete removes a budget.
func (s *BudgetService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete budget %d: %w", id, err)
	}
	return nil
}

// CheckThresholds returns budgets that have reached their notification threshold.
func (s *BudgetService) CheckThresholds(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	budgets, err := s.ListWithProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	var alerts []*domain.Budget
	for _, b := range budgets {
		if b.ShouldNotify() {
			alerts = append(alerts, b)
		}
	}
	return alerts, nil
}

// periodBounds returns the start and end of the current budget period.
func periodBounds(period domain.BudgetPeriod, now time.Time) (from, to time.Time) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch period {
	case domain.BudgetPeriodWeekly:
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		from = today.AddDate(0, 0, -(weekday - 1))
		to = from.AddDate(0, 0, 7)
	case domain.BudgetPeriodMonthly:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 1, 0)
	default:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		to = from.AddDate(0, 1, 0)
	}
	return from, to
}
