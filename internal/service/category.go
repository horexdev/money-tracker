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

// Create adds a new custom category for a user.
func (s *CategoryService) Create(ctx context.Context, userID int64, name, emoji, catType string) (*domain.Category, error) {
	if name == "" {
		return nil, fmt.Errorf("category name is required: %w", domain.ErrInvalidAmount)
	}
	if catType == "" {
		catType = string(domain.CategoryTypeBoth)
	}

	cat, err := s.repo.CreateForUser(ctx, userID, name, emoji, catType)
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
func (s *CategoryService) Update(ctx context.Context, userID, categoryID int64, name, emoji, catType string) (*domain.Category, error) {
	existing, err := s.repo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if existing.IsSystem() {
		return nil, domain.ErrCategorySystemReadOnly
	}
	if existing.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}

	cat, err := s.repo.Update(ctx, userID, categoryID, name, emoji, catType)
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
	if existing.IsSystem() {
		return domain.ErrCategorySystemReadOnly
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
	if !cat.IsSystem() && cat.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, nil
}
