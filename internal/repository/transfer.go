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

// TransferRepository implements service.TransferStorer using pgx + sqlc.
type TransferRepository struct {
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

// NewTransferRepository constructs a TransferRepository.
func NewTransferRepository(pool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{pool: pool, q: sqlcgen.New(pool)}
}

func transferFromRow(r sqlcgen.Transfer, fromName, toName string) *domain.Transfer {
	return &domain.Transfer{
		ID:               r.ID,
		UserID:           r.UserID,
		FromAccountID:    r.FromAccountID,
		ToAccountID:      r.ToAccountID,
		FromAccountName:  fromName,
		ToAccountName:    toName,
		AmountCents:      r.AmountCents,
		FromCurrencyCode: r.FromCurrencyCode,
		ToCurrencyCode:   r.ToCurrencyCode,
		ExchangeRate:     goFloat64(r.ExchangeRate),
		Note:             r.Note,
		CreatedAt:        r.CreatedAt.Time,
	}
}

// Create inserts a new transfer record.
func (r *TransferRepository) Create(ctx context.Context, t *domain.Transfer) (*domain.Transfer, error) {
	row, err := r.q.CreateTransfer(ctx, sqlcgen.CreateTransferParams{
		UserID:           t.UserID,
		FromAccountID:    t.FromAccountID,
		ToAccountID:      t.ToAccountID,
		AmountCents:      t.AmountCents,
		FromCurrencyCode: t.FromCurrencyCode,
		ToCurrencyCode:   t.ToCurrencyCode,
		ExchangeRate:     pgNumeric(t.ExchangeRate),
		Note:             t.Note,
	})
	if err != nil {
		return nil, fmt.Errorf("create transfer: %w", err)
	}
	return transferFromRow(row, t.FromAccountName, t.ToAccountName), nil
}

// GetByID returns a transfer with joined account names.
func (r *TransferRepository) GetByID(ctx context.Context, id, userID int64) (*domain.Transfer, error) {
	row, err := r.q.GetTransferByID(ctx, sqlcgen.GetTransferByIDParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transfer by id: %w", err)
	}
	return &domain.Transfer{
		ID:               row.ID,
		UserID:           row.UserID,
		FromAccountID:    row.FromAccountID,
		ToAccountID:      row.ToAccountID,
		FromAccountName:  row.FromAccountName,
		ToAccountName:    row.ToAccountName,
		AmountCents:      row.AmountCents,
		FromCurrencyCode: row.FromCurrencyCode,
		ToCurrencyCode:   row.ToCurrencyCode,
		ExchangeRate:     goFloat64(row.ExchangeRate),
		Note:             row.Note,
		CreatedAt:        row.CreatedAt.Time,
	}, nil
}

// ListByUser returns paginated transfers for a user.
func (r *TransferRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transfer, error) {
	rows, err := r.q.ListTransfersByUser(ctx, sqlcgen.ListTransfersByUserParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list transfers by user: %w", err)
	}
	out := make([]*domain.Transfer, 0, len(rows))
	for _, row := range rows {
		out = append(out, &domain.Transfer{
			ID:               row.ID,
			UserID:           row.UserID,
			FromAccountID:    row.FromAccountID,
			ToAccountID:      row.ToAccountID,
			FromAccountName:  row.FromAccountName,
			ToAccountName:    row.ToAccountName,
			AmountCents:      row.AmountCents,
			FromCurrencyCode: row.FromCurrencyCode,
			ToCurrencyCode:   row.ToCurrencyCode,
			ExchangeRate:     goFloat64(row.ExchangeRate),
			Note:             row.Note,
			CreatedAt:        row.CreatedAt.Time,
		})
	}
	return out, nil
}

// ListByAccount returns all transfers involving a specific account.
func (r *TransferRepository) ListByAccount(ctx context.Context, userID, accountID int64) ([]*domain.Transfer, error) {
	rows, err := r.q.ListTransfersByAccount(ctx, sqlcgen.ListTransfersByAccountParams{
		UserID:        userID,
		FromAccountID: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("list transfers by account: %w", err)
	}
	out := make([]*domain.Transfer, 0, len(rows))
	for _, row := range rows {
		out = append(out, &domain.Transfer{
			ID:               row.ID,
			UserID:           row.UserID,
			FromAccountID:    row.FromAccountID,
			ToAccountID:      row.ToAccountID,
			FromAccountName:  row.FromAccountName,
			ToAccountName:    row.ToAccountName,
			AmountCents:      row.AmountCents,
			FromCurrencyCode: row.FromCurrencyCode,
			ToCurrencyCode:   row.ToCurrencyCode,
			ExchangeRate:     goFloat64(row.ExchangeRate),
			Note:             row.Note,
			CreatedAt:        row.CreatedAt.Time,
		})
	}
	return out, nil
}

// Count returns the total number of transfers for a user.
func (r *TransferRepository) Count(ctx context.Context, userID int64) (int64, error) {
	n, err := r.q.CountTransfersByUser(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count transfers: %w", err)
	}
	return n, nil
}

// Delete removes a transfer.
func (r *TransferRepository) Delete(ctx context.Context, id, userID int64) error {
	if err := r.q.DeleteTransfer(ctx, sqlcgen.DeleteTransferParams{ID: id, UserID: userID}); err != nil {
		return fmt.Errorf("delete transfer: %w", err)
	}
	return nil
}
