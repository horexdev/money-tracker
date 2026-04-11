package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// BudgetRepository handles persistence of Budget entities.
type BudgetRepository struct {
	q *sqlcgen.Queries
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{q: sqlcgen.New(pool)}
}

// Create inserts a new budget and returns the persisted record.
func (r *BudgetRepository) Create(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	row, err := r.q.CreateBudget(ctx, sqlcgen.CreateBudgetParams{
		UserID:               b.UserID,
		CategoryID:           b.CategoryID,
		LimitCents:           b.LimitCents,
		Period:               string(b.Period),
		CurrencyCode:         b.CurrencyCode,
		NotifyAtPercent:      int32(b.NotifyAtPercent),
		NotificationsEnabled: b.NotificationsEnabled,
	})
	if err != nil {
		return nil, err
	}
	return rowToBudget(row), nil
}

// GetByID returns the budget with the given ID scoped to user.
func (r *BudgetRepository) GetByID(ctx context.Context, id, userID int64) (*domain.Budget, error) {
	row, err := r.q.GetBudgetByID(ctx, sqlcgen.GetBudgetByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return rowToBudget(row), nil
}

// ListByUser returns all budgets for a user with category info.
func (r *BudgetRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.Budget, error) {
	rows, err := r.q.ListBudgetsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	budgets := make([]*domain.Budget, 0, len(rows))
	for _, row := range rows {
		b := &domain.Budget{
			ID:                   row.ID,
			UserID:               row.UserID,
			CategoryID:           row.CategoryID,
			LimitCents:           row.LimitCents,
			Period:               domain.BudgetPeriod(row.Period),
			CurrencyCode:         row.CurrencyCode,
			NotifyAtPercent:      int(row.NotifyAtPercent),
			NotificationsEnabled: row.NotificationsEnabled,
			LastNotifiedPercent:  int(row.LastNotifiedPercent),
			CreatedAt:            goTime(row.CreatedAt),
			UpdatedAt:            goTime(row.UpdatedAt),
			CategoryName:         row.CategoryName,
			CategoryIcon:         row.CategoryIcon,
			CategoryColor:        row.CategoryColor,
		}
		if row.LastNotifiedAt.Valid {
			t := row.LastNotifiedAt.Time
			b.LastNotifiedAt = &t
		}
		budgets = append(budgets, b)
	}
	return budgets, nil
}

// Update modifies an existing budget.
func (r *BudgetRepository) Update(ctx context.Context, b *domain.Budget) (*domain.Budget, error) {
	row, err := r.q.UpdateBudget(ctx, sqlcgen.UpdateBudgetParams{
		ID:                   b.ID,
		UserID:               b.UserID,
		LimitCents:           b.LimitCents,
		Period:               string(b.Period),
		CurrencyCode:         b.CurrencyCode,
		NotifyAtPercent:      int32(b.NotifyAtPercent),
		NotificationsEnabled: b.NotificationsEnabled,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return rowToBudget(row), nil
}

// Delete removes a budget by ID scoped to user.
func (r *BudgetRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.q.DeleteBudget(ctx, sqlcgen.DeleteBudgetParams{ID: id, UserID: userID})
}

// GetByUserCategoryPeriod checks if a budget already exists for the given combination.
func (r *BudgetRepository) GetByUserCategoryPeriod(ctx context.Context, userID, categoryID int64, period string) (*domain.Budget, error) {
	row, err := r.q.GetBudgetByUserCategoryPeriod(ctx, sqlcgen.GetBudgetByUserCategoryPeriodParams{
		UserID:     userID,
		CategoryID: categoryID,
		Period:     period,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return rowToBudget(row), nil
}

// GetSpentInPeriod returns total expense amount for a category in a date range,
// converting all transactions to the target currency via exchange_rate_snapshots.
func (r *BudgetRepository) GetSpentInPeriod(ctx context.Context, userID, categoryID int64, currency string, from, to time.Time) (int64, error) {
	return r.q.GetSpentInPeriod(ctx, sqlcgen.GetSpentInPeriodParams{
		UserID:         userID,
		CategoryID:     categoryID,
		TargetCurrency: currency,
		CreatedAt:      pgTimestamptz(from),
		CreatedAt_2:    pgTimestamptz(to),
	})
}

// UpdateLastNotified sets last_notified_at = now() and records the threshold percent.
func (r *BudgetRepository) UpdateLastNotified(ctx context.Context, id int64, notifiedPercent int) error {
	return r.q.UpdateBudgetLastNotified(ctx, sqlcgen.UpdateBudgetLastNotifiedParams{
		ID:                  id,
		LastNotifiedPercent: int32(notifiedPercent),
	})
}

// ListDistinctUserIDs returns all user IDs that have at least one budget.
func (r *BudgetRepository) ListDistinctUserIDs(ctx context.Context) ([]int64, error) {
	rows, err := r.q.ListDistinctUsersWithBudgets(ctx)
	if err != nil {
		return nil, err
	}
	ids := append(make([]int64, 0, len(rows)), rows...)
	return ids, nil
}

func rowToBudget(row sqlcgen.Budget) *domain.Budget {
	b := &domain.Budget{
		ID:                   row.ID,
		UserID:               row.UserID,
		CategoryID:           row.CategoryID,
		LimitCents:           row.LimitCents,
		Period:               domain.BudgetPeriod(row.Period),
		CurrencyCode:         row.CurrencyCode,
		NotifyAtPercent:      int(row.NotifyAtPercent),
		NotificationsEnabled: row.NotificationsEnabled,
		LastNotifiedPercent:  int(row.LastNotifiedPercent),
		CreatedAt:            goTime(row.CreatedAt),
		UpdatedAt:            goTime(row.UpdatedAt),
	}
	if row.LastNotifiedAt.Valid {
		t := row.LastNotifiedAt.Time
		b.LastNotifiedAt = &t
	}
	return b
}
