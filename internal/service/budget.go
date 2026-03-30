package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// BudgetNotifier sends budget threshold alerts to users.
type BudgetNotifier interface {
	SendBudgetAlert(ctx context.Context, chatID int64, lang, categoryName, currencyCode string, spentPercent int, limitCents, spentCents int64) error
}

// BudgetService handles business logic for budgets.
type BudgetService struct {
	repo     BudgetStorer
	txRepo   TransactionStorer
	userRepo UserStorer
	notifier BudgetNotifier
	log      *slog.Logger
}

func NewBudgetService(repo BudgetStorer, txRepo TransactionStorer, userRepo UserStorer, log *slog.Logger) *BudgetService {
	return &BudgetService{repo: repo, txRepo: txRepo, userRepo: userRepo, log: log}
}

// WithNotifier sets the notifier used by CheckAndNotify.
func (s *BudgetService) WithNotifier(n BudgetNotifier) {
	s.notifier = n
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

// ListForBudget returns expense transactions that contributed to a budget's spending.
func (s *BudgetService) ListForBudget(ctx context.Context, budgetID, userID int64) ([]*domain.Transaction, error) {
	budget, err := s.repo.GetByID(ctx, budgetID, userID)
	if err != nil {
		return nil, fmt.Errorf("get budget %d: %w", budgetID, err)
	}

	from, to := periodBounds(budget.Period, time.Now())
	txs, err := s.txRepo.ListByCategoryPeriod(ctx, userID, budget.CategoryID, from, to)
	if err != nil {
		return nil, fmt.Errorf("list transactions for budget %d: %w", budgetID, err)
	}
	return txs, nil
}

// ListDistinctUserIDs returns all user IDs that have at least one budget.
func (s *BudgetService) ListDistinctUserIDs(ctx context.Context) ([]int64, error) {
	return s.repo.ListDistinctUserIDs(ctx)
}

// CheckAndNotify sends Telegram alerts for budgets that have crossed a threshold
// and have not yet been notified for that threshold in the current period.
func (s *BudgetService) CheckAndNotify(ctx context.Context, userID int64) error {
	if s.notifier == nil {
		return nil
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user %d: %w", userID, err)
	}
	lang := string(user.Language)
	if lang == "" {
		lang = "en"
	}

	budgets, err := s.ListWithProgress(ctx, userID)
	if err != nil {
		return fmt.Errorf("list budgets: %w", err)
	}

	for _, b := range budgets {
		threshold := b.NextAlertThreshold()
		if threshold == 0 {
			continue
		}

		if err := s.notifier.SendBudgetAlert(ctx, userID, lang, b.CategoryName, b.CurrencyCode, threshold, b.LimitCents, b.SpentCents); err != nil {
			s.log.WarnContext(ctx, "failed to send budget alert",
				slog.Int64("user_id", userID),
				slog.Int64("budget_id", b.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		if err := s.repo.UpdateLastNotified(ctx, b.ID, threshold); err != nil {
			s.log.WarnContext(ctx, "failed to update last_notified_at",
				slog.Int64("budget_id", b.ID),
				slog.String("error", err.Error()),
			)
		}
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
