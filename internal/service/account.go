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
	repo    AccountStorer
	exchSvc *ExchangeService
	log     *slog.Logger
}

// NewAccountService constructs an AccountService.
func NewAccountService(repo AccountStorer, exchSvc *ExchangeService, log *slog.Logger) *AccountService {
	return &AccountService{repo: repo, exchSvc: exchSvc, log: log}
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

// List returns all accounts for a user with computed balances converted to each account's currency.
func (s *AccountService) List(ctx context.Context, userID int64) ([]*domain.Account, error) {
	accounts, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	for _, a := range accounts {
		bal, err := s.balanceInCurrency(ctx, a.ID, userID, a.CurrencyCode)
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

// GetByID returns a single account with its balance converted to the account's currency.
func (s *AccountService) GetByID(ctx context.Context, id, userID int64) (*domain.Account, error) {
	a, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	bal, err := s.balanceInCurrency(ctx, a.ID, userID, a.CurrencyCode)
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

// balanceInCurrency returns the account balance converted to targetCurrency.
// Each transaction is first converted to base currency using its exchange_rate_snapshot,
// then the sum is converted from base to targetCurrency via ExchangeService.
// Falls back to raw balance if there are no transactions or exchange rates are unavailable.
func (s *AccountService) balanceInCurrency(ctx context.Context, accountID, userID int64, targetCurrency string) (int64, error) {
	balInBase, err := s.repo.GetBalanceInBase(ctx, accountID, userID)
	if err != nil {
		return 0, err
	}
	if balInBase == 0 {
		return 0, nil
	}

	baseCurrency, err := s.repo.GetBaseCurrency(ctx, accountID, userID)
	if err != nil {
		// No transactions yet — balance is zero.
		return 0, nil
	}

	if baseCurrency == targetCurrency {
		return balInBase, nil
	}

	converted, err := s.exchSvc.Convert(ctx, balInBase, baseCurrency, targetCurrency)
	if err != nil {
		// Exchange rate unavailable — fall back to raw balance to avoid showing 0.
		s.log.WarnContext(ctx, "exchange rate unavailable, falling back to raw balance",
			slog.Int64("account_id", accountID),
			slog.String("base", baseCurrency),
			slog.String("target", targetCurrency),
			slog.String("err", err.Error()),
		)
		return s.repo.GetBalance(ctx, accountID, userID)
	}
	return converted, nil
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
