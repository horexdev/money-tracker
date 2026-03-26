package repository

import (
	"context"
	"database/sql"
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
	rows, err := r.q.ListUserCategories(ctx, userID)
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
// Returns domain.ErrCategoryNotFound if not found.
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

func rowToCategory(row sqlcgen.Category) *domain.Category {
	var userID int64
	if row.UserID.Valid {
		userID = row.UserID.Int64
	}
	return &domain.Category{
		ID:     row.ID,
		UserID: userID,
		Name:   row.Name,
		Emoji:  row.Emoji,
	}
}

// CreateForUser adds a custom category for a specific user.
func (r *CategoryRepository) CreateForUser(ctx context.Context, userID int64, name, emoji string) (*domain.Category, error) {
	row, err := r.q.CreateUserCategory(ctx, sqlcgen.CreateUserCategoryParams{
		UserID: sql.NullInt64{Int64: userID, Valid: true},
		Name:   name,
		Emoji:  emoji,
	})
	if err != nil {
		return nil, err
	}
	return rowToCategory(row), nil
}
