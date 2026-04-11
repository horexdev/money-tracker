package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// CategoryService handles business logic for category management.
type CategoryService struct {
	repo CategoryStorer
	log  *slog.Logger
}

func NewCategoryService(repo CategoryStorer, log *slog.Logger) *CategoryService {
	return &CategoryService{repo: repo, log: log}
}

// ListForUser returns all categories available to a user.
func (s *CategoryService) ListForUser(ctx context.Context, userID int64) ([]*domain.Category, error) {
	cats, err := s.repo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories for user %d: %w", userID, err)
	}
	return cats, nil
}

// ListForUserByType returns categories filtered by transaction type.
func (s *CategoryService) ListForUserByType(ctx context.Context, userID int64, catType string) ([]*domain.Category, error) {
	cats, err := s.repo.ListForUserByType(ctx, userID, catType)
	if err != nil {
		return nil, fmt.Errorf("list categories by type for user %d: %w", userID, err)
	}
	return cats, nil
}

// ListSorted returns categories optionally filtered by type and sorted by name.
// catType filters by category type when non-empty ("expense", "income", "both").
// order controls sort direction: "asc" (default) or "desc".
func (s *CategoryService) ListSorted(ctx context.Context, userID int64, catType, order string) ([]*domain.Category, error) {
	switch order {
	case "", "asc":
		order = "asc"
	case "desc":
		// valid
	default:
		return nil, fmt.Errorf("order %q: %w", order, domain.ErrInvalidSortParam)
	}

	switch catType {
	case "", "expense", "income", "both":
		// valid
	default:
		return nil, fmt.Errorf("type %q: %w", catType, domain.ErrInvalidSortParam)
	}

	cats, err := s.repo.ListSorted(ctx, userID, catType, order)
	if err != nil {
		return nil, fmt.Errorf("list sorted categories for user %d: %w", userID, err)
	}
	return cats, nil
}

// HasCategories reports whether the user has at least one personal category.
func (s *CategoryService) HasCategories(ctx context.Context, userID int64) (bool, error) {
	has, err := s.repo.HasCategories(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("check categories for user %d: %w", userID, err)
	}
	return has, nil
}

// InitDefaultForUser seeds a set of default categories for a new user.
// Seeds are provided by the caller (API layer) with the appropriate locale.
func (s *CategoryService) InitDefaultForUser(ctx context.Context, userID int64, seeds []domain.CategorySeed) error {
	if err := s.repo.BulkCreateForUser(ctx, userID, seeds); err != nil {
		return fmt.Errorf("init default categories for user %d: %w", userID, err)
	}
	s.log.InfoContext(ctx, "default categories seeded",
		slog.Int64("user_id", userID),
		slog.Int("count", len(seeds)),
	)
	return nil
}

// Create adds a new custom category for a user.
func (s *CategoryService) Create(ctx context.Context, userID int64, name, icon, catType, color string) (*domain.Category, error) {

	if name == "" {
		return nil, domain.ErrCategoryNameEmpty
	}
	if catType == "" {
		catType = string(domain.CategoryTypeBoth)
	}
	if color == "" {
		color = "#6366f1"
	}

	cat, err := s.repo.CreateForUser(ctx, userID, name, icon, catType, color)
	if err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	s.log.InfoContext(ctx, "category created",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", cat.ID),
		slog.String("name", name),
	)
	return cat, nil
}

// Update modifies a user-owned category.
func (s *CategoryService) Update(ctx context.Context, userID, categoryID int64, name, icon, catType, color string) (*domain.Category, error) {

	existing, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if existing.IsSystem() || existing.IsProtected {
		return nil, domain.ErrCategoryProtected
	}
	if existing.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}
	if color == "" {
		color = existing.Color
	}

	cat, err := s.repo.Update(ctx, userID, categoryID, name, icon, catType, color)
	if err != nil {
		return nil, fmt.Errorf("update category %d: %w", categoryID, err)
	}

	s.log.InfoContext(ctx, "category updated",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
	)
	return cat, nil
}

// Delete soft-deletes a user-owned category. Returns ErrCategoryInUse if transactions reference it.
func (s *CategoryService) Delete(ctx context.Context, userID, categoryID int64) error {
	existing, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if existing.IsSystem() || existing.IsProtected {
		return domain.ErrCategoryProtected
	}
	if existing.UserID != userID {
		return domain.ErrCategoryNotFound
	}

	count, err := s.repo.CountTransactions(ctx, categoryID)
	if err != nil {
		return fmt.Errorf("count transactions for category %d: %w", categoryID, err)
	}
	if count > 0 {
		return domain.ErrCategoryInUse
	}

	if err := s.repo.SoftDelete(ctx, categoryID, userID); err != nil {
		return fmt.Errorf("delete category %d: %w", categoryID, err)
	}

	s.log.InfoContext(ctx, "category deleted",
		slog.Int64("user_id", userID),
		slog.Int64("category_id", categoryID),
	)
	return nil
}

// GetByID returns a category by ID, validating ownership.
func (s *CategoryService) GetByID(ctx context.Context, userID, categoryID int64) (*domain.Category, error) {
	cat, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		if errors.Is(err, domain.ErrCategoryNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if cat.IsPersonal() && cat.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, nil
}
