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
		UserID:      t.UserID,
		Type:        t.Type,
		AmountCents: t.AmountCents,
		CategoryID:  t.CategoryID,
		Note:        t.Note,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Transaction{
		ID:          row.ID,
		UserID:      row.UserID,
		Type:        row.Type,
		AmountCents: row.AmountCents,
		CategoryID:  row.CategoryID,
		Note:        row.Note,
		CreatedAt:   row.CreatedAt,
	}, nil
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
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			CategoryID:    row.CategoryID,
			CategoryName:  row.CategoryName,
			CategoryEmoji: row.CategoryEmoji,
			Note:          row.Note,
			CreatedAt:     row.CreatedAt,
		})
	}
	return txs, nil
}

// StatsByCategory returns aggregated stats per category for the given period.
func (r *TransactionRepository) StatsByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	rows, err := r.q.GetStatsByCategory(ctx, sqlcgen.GetStatsByCategoryParams{
		UserID:    userID,
		CreatedAt: from,
		EndAt:     to,
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
			TotalCents:    row.TotalCents,
			TxCount:       row.TxCount,
		})
	}
	return stats, nil
}
