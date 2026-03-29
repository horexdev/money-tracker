package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// AccountService handles business logic for user accounts.
type AccountService struct {
	repo AccountStorer
	log  *slog.Logger
}

// NewAccountService constructs an AccountService.
func NewAccountService(repo AccountStorer, log *slog.Logger) *AccountService {
	return &AccountService{repo: repo, log: log}
}

// Create creates a new account. If it is the first account for the user, it is
// automatically made the default.
func (s *AccountService) Create(ctx context.Context, userID int64, name, icon, color string, accType domain.AccountType, currencyCode string, includeInTotal bool) (*domain.Account, error) {
	existing, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	isDefault := len(existing) == 0

	a := &domain.Account{
		UserID:         userID,
		Name:           name,
		Icon:           icon,
		Color:          color,
		Type:           accType,
		CurrencyCode:   currencyCode,
		IsDefault:      isDefault,
		IncludeInTotal: includeInTotal,
	}

	created, err := s.repo.Create(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	s.log.InfoContext(ctx, "account created",
		slog.Int64("user_id", userID),
		slog.Int64("account_id", created.ID),
		slog.String("name", name),
	)
	return created, nil
}

// List returns all accounts for a user with computed balances.
func (s *AccountService) List(ctx context.Context, userID int64) ([]*domain.Account, error) {
	accounts, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	for _, a := range accounts {
		bal, err := s.repo.GetBalance(ctx, a.ID, userID)
		if err != nil {
			s.log.WarnContext(ctx, "failed to get account balance",
				slog.Int64("account_id", a.ID),
				slog.String("err", err.Error()),
			)
			continue
		}
		a.BalanceCents = bal
	}
	return accounts, nil
}

// GetByID returns a single account with its balance.
func (s *AccountService) GetByID(ctx context.Context, id, userID int64) (*domain.Account, error) {
	a, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	bal, err := s.repo.GetBalance(ctx, a.ID, userID)
	if err != nil {
		s.log.WarnContext(ctx, "failed to get account balance",
			slog.Int64("account_id", a.ID),
			slog.String("err", err.Error()),
		)
	} else {
		a.BalanceCents = bal
	}
	return a, nil
}

// Update modifies a non-default account. is_default cannot be changed via Update — use SetDefault.
func (s *AccountService) Update(ctx context.Context, id, userID int64, name, icon, color string, accType domain.AccountType, currencyCode string, includeInTotal bool) (*domain.Account, error) {
	existing, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	existing.Name = name
	existing.Icon = icon
	existing.Color = color
	existing.Type = accType
	existing.CurrencyCode = currencyCode
	existing.IncludeInTotal = includeInTotal

	updated, err := s.repo.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("update account: %w", err)
	}
	return updated, nil
}

// SetDefault makes the given account the default for the user.
func (s *AccountService) SetDefault(ctx context.Context, id, userID int64) (*domain.Account, error) {
	a, err := s.repo.SetDefault(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("set default account: %w", err)
	}
	s.log.InfoContext(ctx, "default account changed",
		slog.Int64("user_id", userID),
		slog.Int64("account_id", id),
	)
	return a, nil
}

// Delete removes an account if it has no transactions.
func (s *AccountService) Delete(ctx context.Context, id, userID int64) error {
	count, err := s.repo.CountTransactions(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("count transactions: %w", err)
	}
	if count > 0 {
		return domain.ErrAccountHasTransactions
	}
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		if errors.Is(err, domain.ErrAccountNotFound) {
			return err
		}
		return fmt.Errorf("delete account: %w", err)
	}
	s.log.InfoContext(ctx, "account deleted",
		slog.Int64("user_id", userID),
		slog.Int64("account_id", id),
	)
	return nil
}

// GetDefault returns the user's default account.
func (s *AccountService) GetDefault(ctx context.Context, userID int64) (*domain.Account, error) {
	a, err := s.repo.GetDefault(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get default account: %w", err)
	}
	return a, nil
}
