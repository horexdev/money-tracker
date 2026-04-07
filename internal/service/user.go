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

// UpdateLanguage changes the user's preferred language.
func (s *UserService) UpdateLanguage(ctx context.Context, id int64, lang string) (*domain.User, error) {
	if !domain.ValidLanguage(lang) {
		return nil, domain.ErrInvalidLanguage
	}
	u, err := s.repo.UpdateLanguage(ctx, id, lang)
	if err != nil {
		return nil, fmt.Errorf("update language for user %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "language updated", slog.Int64("user_id", id), slog.String("language", lang))
	return u, nil
}

// UpdateNotificationPreferences saves the user's notification opt-in settings.
func (s *UserService) UpdateNotificationPreferences(ctx context.Context, id int64, prefs domain.NotificationPrefs) (*domain.User, error) {
	u, err := s.repo.UpdateNotificationPreferences(ctx, id, prefs)
	if err != nil {
		return nil, fmt.Errorf("update notification preferences for user %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "notification preferences updated", slog.Int64("user_id", id))
	return u, nil
}

// ResetData deletes all user-owned data while keeping the user account and settings.
func (s *UserService) ResetData(ctx context.Context, userID int64) error {
	if err := s.repo.ResetData(ctx, userID); err != nil {
		return fmt.Errorf("reset data for user %d: %w", userID, err)
	}
	s.log.InfoContext(ctx, "user data reset", slog.Int64("user_id", userID))
	return nil
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
