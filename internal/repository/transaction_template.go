package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// TransactionTemplateRepository persists user-defined transaction templates.
type TransactionTemplateRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewTransactionTemplateRepository(pool *pgxpool.Pool) *TransactionTemplateRepository {
	return &TransactionTemplateRepository{pool: pool, q: sqlcgen.New(pool)}
}

// Create inserts a new template. sort_order is auto-assigned to MAX+1 per user.
func (r *TransactionTemplateRepository) Create(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	row, err := r.q.CreateTransactionTemplate(ctx, sqlcgen.CreateTransactionTemplateParams{
		UserID:       t.UserID,
		Name:         t.Name,
		Type:         t.Type,
		AmountCents:  t.AmountCents,
		AmountFixed:  t.AmountFixed,
		CategoryID:   t.CategoryID,
		AccountID:    t.AccountID,
		CurrencyCode: t.CurrencyCode,
		Note:         t.Note,
	})
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	// Reload with joined category fields.
	return r.GetByID(ctx, row.ID, t.UserID)
}

// GetByID returns a template scoped to user, with joined category info.
func (r *TransactionTemplateRepository) GetByID(ctx context.Context, id, userID int64) (*domain.TransactionTemplate, error) {
	row, err := r.q.GetTransactionTemplateByID(ctx, sqlcgen.GetTransactionTemplateByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &domain.TransactionTemplate{
		ID:            row.ID,
		UserID:        row.UserID,
		Name:          row.Name,
		Type:          row.Type,
		AmountCents:   row.AmountCents,
		AmountFixed:   row.AmountFixed,
		CategoryID:    row.CategoryID,
		AccountID:     row.AccountID,
		CurrencyCode:  row.CurrencyCode,
		Note:          row.Note,
		SortOrder:     row.SortOrder,
		CreatedAt:     goTime(row.CreatedAt),
		UpdatedAt:     goTime(row.UpdatedAt),
		CategoryName:  row.CategoryName,
		CategoryIcon:  row.CategoryIcon,
		CategoryColor: row.CategoryColor,
	}, nil
}

// ListByUser returns all templates for a user, ordered by sort_order.
func (r *TransactionTemplateRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.TransactionTemplate, error) {
	rows, err := r.q.ListTransactionTemplatesByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	result := make([]*domain.TransactionTemplate, 0, len(rows))
	for _, row := range rows {
		result = append(result, &domain.TransactionTemplate{
			ID:            row.ID,
			UserID:        row.UserID,
			Name:          row.Name,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			AmountFixed:   row.AmountFixed,
			CategoryID:    row.CategoryID,
			AccountID:     row.AccountID,
			CurrencyCode:  row.CurrencyCode,
			Note:          row.Note,
			SortOrder:     row.SortOrder,
			CreatedAt:     goTime(row.CreatedAt),
			UpdatedAt:     goTime(row.UpdatedAt),
			CategoryName:  row.CategoryName,
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
		})
	}
	return result, nil
}

// Update modifies a template. sort_order is preserved (use Reorder to change it).
func (r *TransactionTemplateRepository) Update(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	_, err := r.q.UpdateTransactionTemplate(ctx, sqlcgen.UpdateTransactionTemplateParams{
		ID:           t.ID,
		UserID:       t.UserID,
		Name:         t.Name,
		Type:         t.Type,
		AmountCents:  t.AmountCents,
		AmountFixed:  t.AmountFixed,
		CategoryID:   t.CategoryID,
		AccountID:    t.AccountID,
		CurrencyCode: t.CurrencyCode,
		Note:         t.Note,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, fmt.Errorf("update template: %w", err)
	}
	return r.GetByID(ctx, t.ID, t.UserID)
}

// Delete removes a template scoped to user. Returns ErrTemplateNotFound if no row matched.
func (r *TransactionTemplateRepository) Delete(ctx context.Context, id, userID int64) error {
	rows, err := r.q.DeleteTransactionTemplate(ctx, sqlcgen.DeleteTransactionTemplateParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("delete template: %w", err)
	}
	if rows == 0 {
		return domain.ErrTemplateNotFound
	}
	return nil
}

// Reorder applies a new sort_order to the given user's templates in a single transaction.
// IDs not owned by the user are silently ignored (rowsAffected = 0).
func (r *TransactionTemplateRepository) Reorder(ctx context.Context, userID int64, orderedIDs []int64) error {
	if len(orderedIDs) == 0 {
		return nil
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.q.WithTx(tx)
	for i, id := range orderedIDs {
		if _, err := q.UpdateTransactionTemplateSortOrder(ctx, sqlcgen.UpdateTransactionTemplateSortOrderParams{
			ID:        id,
			UserID:    userID,
			SortOrder: int32(i),
		}); err != nil {
			return fmt.Errorf("update sort_order id=%d: %w", id, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// CountByAccount returns how many templates reference the given account.
func (r *TransactionTemplateRepository) CountByAccount(ctx context.Context, accountID int64) (int64, error) {
	n, err := r.q.CountTransactionTemplatesByAccount(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("count templates by account: %w", err)
	}
	return n, nil
}

// CountByCategory returns how many templates reference the given category.
func (r *TransactionTemplateRepository) CountByCategory(ctx context.Context, categoryID int64) (int64, error) {
	n, err := r.q.CountTransactionTemplatesByCategory(ctx, categoryID)
	if err != nil {
		return 0, fmt.Errorf("count templates by category: %w", err)
	}
	return n, nil
}
