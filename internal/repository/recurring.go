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

// RecurringRepository handles persistence of RecurringTransaction entities.
type RecurringRepository struct {
	q *sqlcgen.Queries
}

func NewRecurringRepository(pool *pgxpool.Pool) *RecurringRepository {
	return &RecurringRepository{q: sqlcgen.New(pool)}
}

// Create inserts a new recurring transaction.
func (r *RecurringRepository) Create(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	row, err := r.q.CreateRecurring(ctx, sqlcgen.CreateRecurringParams{
		UserID:       rt.UserID,
		CategoryID:   rt.CategoryID,
		Type:         rt.Type,
		AmountCents:  rt.AmountCents,
		CurrencyCode: rt.CurrencyCode,
		Note:         rt.Note,
		Frequency:    string(rt.Frequency),
		NextRunAt:    pgTimestamptz(rt.NextRunAt),
	})
	if err != nil {
		return nil, err
	}
	return rowToRecurring(row), nil
}

// GetByID returns a recurring transaction scoped to user.
func (r *RecurringRepository) GetByID(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	row, err := r.q.GetRecurringByID(ctx, sqlcgen.GetRecurringByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecurringNotFound
		}
		return nil, err
	}
	return rowToRecurring(row), nil
}

// ListByUser returns all recurring transactions for a user with category info.
func (r *RecurringRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.RecurringTransaction, error) {
	rows, err := r.q.ListRecurringByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]*domain.RecurringTransaction, 0, len(rows))
	for _, row := range rows {
		result = append(result, &domain.RecurringTransaction{
			ID:            row.ID,
			UserID:        row.UserID,
			CategoryID:    row.CategoryID,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			CurrencyCode:  row.CurrencyCode,
			Note:          row.Note,
			Frequency:     domain.Frequency(row.Frequency),
			NextRunAt:     goTime(row.NextRunAt),
			IsActive:      row.IsActive,
			CreatedAt:     goTime(row.CreatedAt),
			UpdatedAt:     goTime(row.UpdatedAt),
			CategoryName:  row.CategoryName,
			CategoryEmoji: row.CategoryEmoji,
			CategoryColor: row.CategoryColor,
		})
	}
	return result, nil
}

// Update modifies an existing recurring transaction.
func (r *RecurringRepository) Update(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	row, err := r.q.UpdateRecurring(ctx, sqlcgen.UpdateRecurringParams{
		ID:           rt.ID,
		UserID:       rt.UserID,
		CategoryID:   rt.CategoryID,
		Type:         rt.Type,
		AmountCents:  rt.AmountCents,
		CurrencyCode: rt.CurrencyCode,
		Note:         rt.Note,
		Frequency:    string(rt.Frequency),
		NextRunAt:    pgTimestamptz(rt.NextRunAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecurringNotFound
		}
		return nil, err
	}
	return rowToRecurring(row), nil
}

// ToggleActive flips the is_active flag.
func (r *RecurringRepository) ToggleActive(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	row, err := r.q.ToggleRecurringActive(ctx, sqlcgen.ToggleRecurringActiveParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRecurringNotFound
		}
		return nil, err
	}
	return rowToRecurring(row), nil
}

// Delete removes a recurring transaction.
func (r *RecurringRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.q.DeleteRecurring(ctx, sqlcgen.DeleteRecurringParams{ID: id, UserID: userID})
}

// GetDue returns all active recurring transactions due before the given time.
func (r *RecurringRepository) GetDue(ctx context.Context, before time.Time) ([]*domain.RecurringTransaction, error) {
	rows, err := r.q.GetDueRecurring(ctx, pgTimestamptz(before))
	if err != nil {
		return nil, err
	}
	result := make([]*domain.RecurringTransaction, 0, len(rows))
	for _, row := range rows {
		result = append(result, rowToRecurring(row))
	}
	return result, nil
}

// UpdateNextRun sets the next execution time for a recurring transaction.
func (r *RecurringRepository) UpdateNextRun(ctx context.Context, id int64, nextRun time.Time) error {
	return r.q.UpdateRecurringNextRun(ctx, sqlcgen.UpdateRecurringNextRunParams{
		ID:        id,
		NextRunAt: pgTimestamptz(nextRun),
	})
}

func rowToRecurring(row sqlcgen.RecurringTransaction) *domain.RecurringTransaction {
	return &domain.RecurringTransaction{
		ID:           row.ID,
		UserID:       row.UserID,
		CategoryID:   row.CategoryID,
		Type:         row.Type,
		AmountCents:  row.AmountCents,
		CurrencyCode: row.CurrencyCode,
		Note:         row.Note,
		Frequency:    domain.Frequency(row.Frequency),
		NextRunAt:    goTime(row.NextRunAt),
		IsActive:     row.IsActive,
		CreatedAt:    goTime(row.CreatedAt),
		UpdatedAt:    goTime(row.UpdatedAt),
	}
}
