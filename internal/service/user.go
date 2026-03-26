package service

import (
	"context"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository"
)

// UserService handles user registration and lookup.
type UserService struct {
	repo *repository.UserRepository
	log  *slog.Logger
}

func NewUserService(repo *repository.UserRepository, log *slog.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

// Upsert creates or updates a user record from Telegram user data.
func (s *UserService) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u.CurrencyCode == "" {
		u.CurrencyCode = "USD"
	}
	result, err := s.repo.Upsert(ctx, u)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to upsert user",
			slog.Int64("user_id", u.ID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	s.log.InfoContext(ctx, "user upserted", slog.Int64("user_id", result.ID))
	return result, nil
}

// GetByID returns the user with the given ID.
func (s *UserService) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}
