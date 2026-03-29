package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// CategoryRepository handles persistence of Category entities.
type CategoryRepository struct {
	q *sqlcgen.Queries
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{q: sqlcgen.New(pool)}
}

// ListForUser returns all system categories plus user-specific ones.
func (r *CategoryRepository) ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error) {
	rows, err := r.q.ListUserCategories(ctx, pgInt8(userID))
	if err != nil {
		return nil, err
	}
	cats := make([]*domain.Category, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, rowToCategory(row))
	}
	return cats, nil
}

// ListForUserByType returns categories filtered by type (expense/income/both).
func (r *CategoryRepository) ListForUserByType(ctx context.Context, userID int64, catType string) ([]*domain.Category, error) {
	rows, err := r.q.ListUserCategoriesByType(ctx, sqlcgen.ListUserCategoriesByTypeParams{
		UserID: pgInt8(userID),
		Type:   catType,
	})
	if err != nil {
		return nil, err
	}
	cats := make([]*domain.Category, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, rowToCategory(row))
	}
	return cats, nil
}

// GetByID returns the category with the given ID.
func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	row, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategory(row), nil
}

// CreateForUser adds a custom category for a specific user.
func (r *CategoryRepository) CreateForUser(ctx context.Context, userID int64, name, emoji, catType, color string) (*domain.Category, error) {
	row, err := r.q.CreateUserCategory(ctx, sqlcgen.CreateUserCategoryParams{
		UserID: pgInt8(userID),
		Name:   name,
		Emoji:  emoji,
		Type:   catType,
		Color:  color,
	})
	if err != nil {
		return nil, err
	}
	return rowToCategory(row), nil
}

// Update modifies an existing category (must be user-owned).
func (r *CategoryRepository) Update(ctx context.Context, userID, id int64, name, emoji, catType, color string) (*domain.Category, error) {
	row, err := r.q.UpdateCategory(ctx, sqlcgen.UpdateCategoryParams{
		ID:     id,
		UserID: pgInt8(userID),
		Name:   name,
		Emoji:  emoji,
		Type:   catType,
		Color:  color,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategory(row), nil
}

// SoftDelete marks a category as deleted (must be user-owned).
func (r *CategoryRepository) SoftDelete(ctx context.Context, id, userID int64) error {
	return r.q.SoftDeleteCategory(ctx, sqlcgen.SoftDeleteCategoryParams{
		ID:     id,
		UserID: pgInt8(userID),
	})
}

// CountTransactions returns the number of transactions referencing a category.
func (r *CategoryRepository) CountTransactions(ctx context.Context, categoryID int64) (int64, error) {
	return r.q.CountTransactionsByCategory(ctx, categoryID)
}

func rowToCategory(row sqlcgen.Category) *domain.Category {
	cat := &domain.Category{
		ID:        row.ID,
		UserID:    goInt64(row.UserID),
		Name:      row.Name,
		Emoji:     row.Emoji,
		Type:      domain.CategoryType(row.Type),
		Color:     row.Color,
		UpdatedAt: goTime(row.UpdatedAt),
		DeletedAt: goTimePtr(row.DeletedAt),
	}
	return cat
}
