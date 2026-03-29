package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// TransactionRepository handles persistence of Transaction entities.
type TransactionRepository struct {
	q *sqlcgen.Queries
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{q: sqlcgen.New(pool)}
}

// Create inserts a new transaction and returns the persisted record.
func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	row, err := r.q.CreateTransaction(ctx, sqlcgen.CreateTransactionParams{
		UserID:                 t.UserID,
		Type:                   t.Type,
		AmountCents:            t.AmountCents,
		CategoryID:             t.CategoryID,
		Note:                   t.Note,
		CurrencyCode:           t.CurrencyCode,
		ExchangeRateSnapshot:   pgNumeric(t.ExchangeRateSnapshot),
		BaseCurrencyAtCreation: t.BaseCurrencyAtCreation,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Transaction{
		ID:                     row.ID,
		UserID:                 row.UserID,
		Type:                   row.Type,
		AmountCents:            row.AmountCents,
		CategoryID:             row.CategoryID,
		Note:                   row.Note,
		CurrencyCode:           row.CurrencyCode,
		ExchangeRateSnapshot:   goFloat64(row.ExchangeRateSnapshot),
		BaseCurrencyAtCreation: row.BaseCurrencyAtCreation,
		CreatedAt:              goTime(row.CreatedAt),
	}, nil
}

// Delete removes a transaction by ID, scoped to the owning user.
func (r *TransactionRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.q.DeleteTransaction(ctx, sqlcgen.DeleteTransactionParams{ID: id, UserID: userID})
}

// GetBalance returns total income and total expense for a user.
func (r *TransactionRepository) GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error) {
	row, err := r.q.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return row.TotalIncome, row.TotalExpense, nil
}

// List returns paginated transactions for a user with category info.
func (r *TransactionRepository) List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactions(ctx, sqlcgen.ListTransactionsParams{
		UserID: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	txs := make([]*domain.Transaction, 0, len(rows))
	for _, row := range rows {
		txs = append(txs, &domain.Transaction{
			ID:                     row.ID,
			UserID:                 row.UserID,
			Type:                   row.Type,
			AmountCents:            row.AmountCents,
			CategoryID:             row.CategoryID,
			CategoryName:           row.CategoryName,
			CategoryEmoji:          row.CategoryEmoji,
			Note:                   row.Note,
			CurrencyCode:           row.CurrencyCode,
			ExchangeRateSnapshot:   goFloat64(row.ExchangeRateSnapshot),
			BaseCurrencyAtCreation: row.BaseCurrencyAtCreation,
			CreatedAt:              goTime(row.CreatedAt),
		})
	}
	return txs, nil
}

// Count returns the total number of transactions for a user.
func (r *TransactionRepository) Count(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountUserTransactions(ctx, userID)
}

// GetTotalInBaseCurrency returns the net balance in base currency using exchange_rate_snapshot.
func (r *TransactionRepository) GetTotalInBaseCurrency(ctx context.Context, userID int64) (int64, error) {
	return r.q.GetTotalInBaseCurrency(ctx, userID)
}

// GetBalanceByCurrency returns per-currency income/expense totals for a user.
func (r *TransactionRepository) GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error) {
	rows, err := r.q.GetBalanceByCurrency(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.BalanceByCurrency, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.BalanceByCurrency{
			CurrencyCode: row.CurrencyCode,
			IncomeCents:  row.TotalIncome,
			ExpenseCents: row.TotalExpense,
		})
	}
	return result, nil
}

// StatsByCategory returns aggregated stats per category for the given period.
func (r *TransactionRepository) StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	rows, err := r.q.GetStatsByCategory(ctx, sqlcgen.GetStatsByCategoryParams{
		UserID:      userID,
		CreatedAt:   pgTimestamptz(from),
		CreatedAt_2: pgTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}

	stats := make([]domain.CategoryStat, 0, len(rows))
	for _, row := range rows {
		stats = append(stats, domain.CategoryStat{
			CategoryName:  row.CategoryName,
			CategoryEmoji: row.CategoryEmoji,
			Type:          row.Type,
			CurrencyCode:  row.CurrencyCode,
			TotalCents:    row.TotalCents,
			TxCount:       row.TxCount,
		})
	}
	return stats, nil
}
