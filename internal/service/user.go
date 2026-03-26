package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// UserService handles user registration and lookup.
type UserService struct {
	repo UserStorer
	log  *slog.Logger
}

func NewUserService(repo UserStorer, log *slog.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

// Upsert creates or updates a user record from Telegram user data.
func (s *UserService) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u.CurrencyCode == "" {
		u.CurrencyCode = "USD"
	}
	result, err := s.repo.Upsert(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("upsert user %d: %w", u.ID, err)
	}
	s.log.InfoContext(ctx, "user upserted", slog.Int64("user_id", result.ID))
	return result, nil
}

// GetByID returns the user with the given ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user %d: %w", id, err)
	}
	return u, nil
}
