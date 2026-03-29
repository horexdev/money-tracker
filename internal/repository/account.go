package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// AccountRepository implements service.AccountStorer using pgx + sqlc.
type AccountRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewAccountRepository constructs an AccountRepository.
func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool, q: sqlcgen.New(pool)}
}

func accountFromRow(r sqlcgen.Account) *domain.Account {
	return &domain.Account{
		ID:             r.ID,
		UserID:         r.UserID,
		Name:           r.Name,
		Icon:           r.Icon,
		Color:          r.Color,
		Type:           domain.AccountType(r.Type),
		CurrencyCode:   r.CurrencyCode,
		IsDefault:      r.IsDefault,
		IncludeInTotal: r.IncludeInTotal,
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}
}

// Create inserts a new account.
func (r *AccountRepository) Create(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	row, err := r.q.CreateAccount(ctx, sqlcgen.CreateAccountParams{
		UserID:         a.UserID,
		Name:           a.Name,
		Icon:           a.Icon,
		Color:          a.Color,
		Type:           sqlcgen.AccountType(a.Type),
		CurrencyCode:   a.CurrencyCode,
		IsDefault:      a.IsDefault,
		IncludeInTotal: a.IncludeInTotal,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	return accountFromRow(row), nil
}

// GetByID returns an account by ID, scoped to the user.
func (r *AccountRepository) GetByID(ctx context.Context, id, userID int64) (*domain.Account, error) {
	row, err := r.q.GetAccountByID(ctx, sqlcgen.GetAccountByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("get account by id: %w", err)
	}
	return accountFromRow(row), nil
}

// GetDefault returns the default account for a user.
func (r *AccountRepository) GetDefault(ctx context.Context, userID int64) (*domain.Account, error) {
	row, err := r.q.GetDefaultAccount(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("get default account: %w", err)
	}
	return accountFromRow(row), nil
}

// ListByUser returns all accounts for a user.
func (r *AccountRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.Account, error) {
	rows, err := r.q.ListAccountsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by user: %w", err)
	}
	out := make([]*domain.Account, 0, len(rows))
	for _, row := range rows {
		out = append(out, accountFromRow(row))
	}
	return out, nil
}

// Update updates mutable account fields.
func (r *AccountRepository) Update(ctx context.Context, a *domain.Account) (*domain.Account, error) {
	row, err := r.q.UpdateAccount(ctx, sqlcgen.UpdateAccountParams{
		ID:             a.ID,
		UserID:         a.UserID,
		Name:           a.Name,
		Icon:           a.Icon,
		Color:          a.Color,
		Type:           sqlcgen.AccountType(a.Type),
		CurrencyCode:   a.CurrencyCode,
		IncludeInTotal: a.IncludeInTotal,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("update account: %w", err)
	}
	return accountFromRow(row), nil
}

// SetDefault clears the current default and marks the given account as default.
func (r *AccountRepository) SetDefault(ctx context.Context, id, userID int64) (*domain.Account, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.q.WithTx(tx)

	if err := q.ClearDefaultAccounts(ctx, userID); err != nil {
		return nil, fmt.Errorf("clear default accounts: %w", err)
	}

	row, err := q.SetAccountDefault(ctx, sqlcgen.SetAccountDefaultParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAccountNotFound
		}
		return nil, fmt.Errorf("set account default: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	return accountFromRow(row), nil
}

// Delete removes an account.
func (r *AccountRepository) Delete(ctx context.Context, id, userID int64) error {
	if err := r.q.DeleteAccount(ctx, sqlcgen.DeleteAccountParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("delete account: %w", err)
	}
	return nil
}

// CountTransactions returns how many transactions are linked to the account.
func (r *AccountRepository) CountTransactions(ctx context.Context, accountID, userID int64) (int64, error) {
	n, err := r.q.CountAccountTransactions(ctx, sqlcgen.CountAccountTransactionsParams{
		AccountID: accountID,
		UserID:    userID,
	})
	if err != nil {
		return 0, fmt.Errorf("count account transactions: %w", err)
	}
	return n, nil
}

// GetBalance returns the net balance (income - expense) for an account.
func (r *AccountRepository) GetBalance(ctx context.Context, accountID, userID int64) (int64, error) {
	cents, err := r.q.GetAccountBalance(ctx, sqlcgen.GetAccountBalanceParams{
		AccountID: accountID,
		UserID:    userID,
	})
	if err != nil {
		return 0, fmt.Errorf("get account balance: %w", err)
	}
	return cents, nil
}
