package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

// UpdateCurrency changes the user's preferred currency code.
func (s *UserService) UpdateCurrency(ctx context.Context, id int64, code string) (*domain.User, error) {
	if !domain.ValidCurrency(code) {
		return nil, domain.ErrInvalidCurrency
	}
	u, err := s.repo.UpdateCurrency(ctx, id, code)
	if err != nil {
		return nil, fmt.Errorf("update currency for user %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "currency updated", slog.Int64("user_id", id), slog.String("currency", code))
	return u, nil
}

// UpdateDisplayCurrencies sets the user's display currencies (max 3).
func (s *UserService) UpdateDisplayCurrencies(ctx context.Context, id int64, codes []string) (*domain.User, error) {
	if len(codes) > 3 {
		return nil, domain.ErrTooManyDisplayCurrencies
	}
	for _, c := range codes {
		if !domain.ValidCurrency(c) {
			return nil, domain.ErrInvalidCurrency
		}
	}
	csv := strings.Join(codes, ",")
	u, err := s.repo.UpdateDisplayCurrencies(ctx, id, csv)
	if err != nil {
		return nil, fmt.Errorf("update display currencies for user %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "display currencies updated", slog.Int64("user_id", id), slog.String("currencies", csv))
	return u, nil
}
