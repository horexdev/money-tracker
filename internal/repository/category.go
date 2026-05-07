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

// ListForUser returns all personal categories for the user (excludes transfer/adjustment).
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

// ListSorted returns categories sorted by the given order.
// catType filters by category type when non-empty.
// order is "asc"/"desc" for name sorting, or "frequency" for transaction-count
// descending with name-asc tiebreaker.
func (r *CategoryRepository) ListSorted(ctx context.Context, userID int64, catType, order string) ([]*domain.Category, error) {
	var (
		rows []sqlcgen.Category
		err  error
	)

	switch {
	case order == "frequency" && catType != "":
		rows, err = r.q.ListUserCategoriesByFrequencyAndType(ctx, sqlcgen.ListUserCategoriesByFrequencyAndTypeParams{
			UserID: userID,
			Type:   catType,
		})
	case order == "frequency":
		rows, err = r.q.ListUserCategoriesByFrequency(ctx, userID)
	case catType != "" && order == "desc":
		rows, err = r.q.ListUserCategoriesByTypeFilterDesc(ctx, sqlcgen.ListUserCategoriesByTypeFilterDescParams{
			UserID: pgInt8(userID),
			Type:   catType,
		})
	case catType != "":
		rows, err = r.q.ListUserCategoriesByTypeFilterAsc(ctx, sqlcgen.ListUserCategoriesByTypeFilterAscParams{
			UserID: pgInt8(userID),
			Type:   catType,
		})
	case order == "desc":
		rows, err = r.q.ListUserCategoriesByNameDesc(ctx, pgInt8(userID))
	default:
		rows, err = r.q.ListUserCategoriesByNameAsc(ctx, pgInt8(userID))
	}

	if err != nil {
		return nil, err
	}
	cats := make([]*domain.Category, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, rowToCategory(row))
	}
	return cats, nil
}

// HasCategories returns true if the user has at least one personal (non-protected) category.
func (r *CategoryRepository) HasCategories(ctx context.Context, userID int64) (bool, error) {
	return r.q.HasUserCategories(ctx, pgInt8(userID))
}

// GetByName returns a category by name for a user (includes system/infrastructure categories).
func (r *CategoryRepository) GetByName(ctx context.Context, userID int64, name string) (*domain.Category, error) {
	row, err := r.q.GetCategoryByName(ctx, sqlcgen.GetCategoryByNameParams{
		UserID: pgInt8(userID),
		Lower:  name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategory(row), nil
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

// GetSystemCategoryByType returns the system (user_id IS NULL) category of the given type.
// Use this for infrastructure categories like "transfer" and "adjustment" instead of GetByName,
// to avoid collisions with user-owned categories of the same name.
func (r *CategoryRepository) GetSystemCategoryByType(ctx context.Context, catType string) (*domain.Category, error) {
	row, err := r.q.GetSystemCategoryByType(ctx, catType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategory(row), nil
}

// GetBySavingsType returns the user's savings category by type (not by name).
func (r *CategoryRepository) GetBySavingsType(ctx context.Context, userID int64) (*domain.Category, error) {
	row, err := r.q.GetCategoryByTypeForUser(ctx, sqlcgen.GetCategoryByTypeForUserParams{
		UserID: pgInt8(userID),
		Type:   string(domain.CategoryTypeSavings),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCategoryNotFound
		}
		return nil, err
	}
	return rowToCategory(row), nil
}

// CreateForUser adds a custom category for a specific user.
func (r *CategoryRepository) CreateForUser(ctx context.Context, userID int64, name, icon, catType, color string) (*domain.Category, error) {
	row, err := r.q.CreateUserCategory(ctx, sqlcgen.CreateUserCategoryParams{
		UserID: pgInt8(userID),
		Name:   name,
		Icon:   icon,
		Type:   catType,
		Color:  color,
	})
	if err != nil {
		return nil, err
	}
	return rowToCategory(row), nil
}

// BulkCreateForUser inserts a set of default categories for the user, ignoring conflicts.
func (r *CategoryRepository) BulkCreateForUser(ctx context.Context, userID int64, seeds []domain.CategorySeed) error {
	for _, s := range seeds {
		_, err := r.q.CreateUserCategory(ctx, sqlcgen.CreateUserCategoryParams{
			UserID: pgInt8(userID),
			Name:   s.Name,
			Icon:   s.Icon,
			Type:   string(s.Type),
			Color:  s.Color,
		})
		if err != nil {
			// ON CONFLICT is not part of CreateUserCategory; skip duplicates gracefully.
			// A unique violation (23505) means the user already has this category — not an error.
			if isDuplicateError(err) {
				continue
			}
			return err
		}
	}
	return nil
}

// Update modifies an existing category (must be user-owned).
func (r *CategoryRepository) Update(ctx context.Context, userID, id int64, name, icon, catType, color string) (*domain.Category, error) {
	row, err := r.q.UpdateCategory(ctx, sqlcgen.UpdateCategoryParams{
		ID:     id,
		UserID: pgInt8(userID),
		Name:   name,
		Icon:   icon,
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
		ID:          row.ID,
		UserID:      goInt64(row.UserID),
		Name:        row.Name,
		Icon:        row.Icon,
		Type:        domain.CategoryType(row.Type),
		Color:       row.Color,
		IsProtected: row.IsProtected,
		UpdatedAt:   goTime(row.UpdatedAt),
		DeletedAt:   goTimePtr(row.DeletedAt),
	}
	return cat
}
