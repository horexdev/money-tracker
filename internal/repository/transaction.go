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
		UserID:       t.UserID,
		Type:         t.Type,
		AmountCents:  t.AmountCents,
		CategoryID:   t.CategoryID,
		Note:         t.Note,
		CurrencyCode: t.CurrencyCode,
		AccountID:    t.AccountID,
		SnapshotDate: pgDate(&t.SnapshotDate),
	})
	if err != nil {
		return nil, err
	}
	return rowToTransaction(row), nil
}

// CreateWithDate inserts a new transaction with an explicit created_at timestamp.
func (r *TransactionRepository) CreateWithDate(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	row, err := r.q.CreateTransactionWithDate(ctx, sqlcgen.CreateTransactionWithDateParams{
		UserID:       t.UserID,
		Type:         t.Type,
		AmountCents:  t.AmountCents,
		CategoryID:   t.CategoryID,
		Note:         t.Note,
		CurrencyCode: t.CurrencyCode,
		CreatedAt:    pgTimestamptz(t.CreatedAt),
		AccountID:    t.AccountID,
		SnapshotDate: pgDate(&t.SnapshotDate),
	})
	if err != nil {
		return nil, err
	}
	return rowToTransaction(row), nil
}

func rowToTransaction(row sqlcgen.Transaction) *domain.Transaction {
	return &domain.Transaction{
		ID:           row.ID,
		UserID:       row.UserID,
		Type:         row.Type,
		AmountCents:  row.AmountCents,
		CategoryID:   row.CategoryID,
		Note:         row.Note,
		CurrencyCode: row.CurrencyCode,
		CreatedAt:    goTime(row.CreatedAt),
		AccountID:    row.AccountID,
		SnapshotDate: goDateValue(row.SnapshotDate),
		IsAdjustment: row.IsAdjustment,
	}
}

// CreateAdjustment inserts a balance-adjustment transaction (is_adjustment = true).
// The transaction affects account balance but is excluded from history and statistics.
func (r *TransactionRepository) CreateAdjustment(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	row, err := r.q.CreateAdjustmentTransaction(ctx, sqlcgen.CreateAdjustmentTransactionParams{
		UserID:       t.UserID,
		Type:         t.Type,
		AmountCents:  t.AmountCents,
		CategoryID:   t.CategoryID,
		Note:         t.Note,
		CurrencyCode: t.CurrencyCode,
		AccountID:    t.AccountID,
		SnapshotDate: pgDate(&t.SnapshotDate),
	})
	if err != nil {
		return nil, err
	}
	return rowToTransaction(row), nil
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
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			CategoryID:    row.CategoryID,
			CategoryName:  row.CategoryName,
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Note:          row.Note,
			CurrencyCode:  row.CurrencyCode,
			CreatedAt:     goTime(row.CreatedAt),
		})
	}
	return txs, nil
}

// Count returns the total number of transactions for a user.
func (r *TransactionRepository) Count(ctx context.Context, userID int64) (int64, error) {
	return r.q.CountUserTransactions(ctx, userID)
}

// GetTotalInBaseCurrency returns the net balance converted to targetCurrency using exchange_rate_snapshots.
func (r *TransactionRepository) GetTotalInBaseCurrency(ctx context.Context, userID int64, targetCurrency string) (int64, error) {
	return r.q.GetTotalInBaseCurrency(ctx, sqlcgen.GetTotalInBaseCurrencyParams{
		UserID:         userID,
		TargetCurrency: targetCurrency,
	})
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

// ListByCategoryPeriod returns expense transactions for a specific category within a time range.
func (r *TransactionRepository) ListByCategoryPeriod(ctx context.Context, userID, categoryID int64, from, to time.Time) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactionsByCategoryPeriod(ctx, sqlcgen.ListTransactionsByCategoryPeriodParams{
		UserID:      userID,
		CategoryID:  categoryID,
		CreatedAt:   pgTimestamptz(from),
		CreatedAt_2: pgTimestamptz(to),
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
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Note:          row.Note,
			CurrencyCode:  row.CurrencyCode,
			CreatedAt:     goTime(row.CreatedAt),
		})
	}
	return txs, nil
}

// Update modifies amount, category, note, and date of an existing transaction.
func (r *TransactionRepository) Update(ctx context.Context, t *domain.Transaction) (*domain.Transaction, error) {
	row, err := r.q.UpdateTransaction(ctx, sqlcgen.UpdateTransactionParams{
		ID:          t.ID,
		UserID:      t.UserID,
		AmountCents: t.AmountCents,
		CategoryID:  t.CategoryID,
		Note:        t.Note,
		CreatedAt:   pgTimestamptz(t.CreatedAt),
	})
	if err != nil {
		return nil, err
	}
	return &domain.Transaction{
		ID:           row.ID,
		UserID:       row.UserID,
		Type:         row.Type,
		AmountCents:  row.AmountCents,
		CategoryID:   row.CategoryID,
		Note:         row.Note,
		CurrencyCode: row.CurrencyCode,
		CreatedAt:    goTime(row.CreatedAt),
	}, nil
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
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Type:          row.Type,
			CurrencyCode:  row.CurrencyCode,
			TotalCents:    row.TotalCents,
			TxCount:       row.TxCount,
		})
	}
	return stats, nil
}

// ListByAccount returns paginated transactions for a specific account.
func (r *TransactionRepository) ListByAccount(ctx context.Context, userID, accountID int64, limit, offset int) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactionsByAccount(ctx, sqlcgen.ListTransactionsByAccountParams{
		UserID:    userID,
		AccountID: accountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
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
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Note:          row.Note,
			CurrencyCode:  row.CurrencyCode,
			CreatedAt:     goTime(row.CreatedAt),
		})
	}
	return txs, nil
}

// CountByAccount returns total transactions count for a specific account.
func (r *TransactionRepository) CountByAccount(ctx context.Context, userID, accountID int64) (int64, error) {
	return r.q.CountUserTransactionsByAccount(ctx, sqlcgen.CountUserTransactionsByAccountParams{
		UserID:    userID,
		AccountID: accountID,
	})
}

// StatsByCategoryAndAccount returns stats filtered by account.
func (r *TransactionRepository) StatsByCategoryAndAccount(ctx context.Context, userID, accountID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	rows, err := r.q.GetStatsByCategoryAndAccount(ctx, sqlcgen.GetStatsByCategoryAndAccountParams{
		UserID:      userID,
		AccountID:   accountID,
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
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Type:          row.Type,
			CurrencyCode:  row.CurrencyCode,
			TotalCents:    row.TotalCents,
			TxCount:       row.TxCount,
		})
	}
	return stats, nil
}

// ListWithDateRange returns paginated transactions filtered by an optional date range.
func (r *TransactionRepository) ListWithDateRange(ctx context.Context, userID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactionsWithDateRange(ctx, sqlcgen.ListTransactionsWithDateRangeParams{
		UserID:  userID,
		Limit:   int32(limit),
		Offset:  int32(offset),
		Column4: pgOptionalTimestamptz(from),
		Column5: pgOptionalTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	txs := make([]*domain.Transaction, 0, len(rows))
	for _, row := range rows {
		var accountName string
		if row.AccountName.Valid {
			accountName = row.AccountName.String
		}
		txs = append(txs, &domain.Transaction{
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			CategoryID:    row.CategoryID,
			CategoryName:  row.CategoryName,
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Note:          row.Note,
			CurrencyCode:  row.CurrencyCode,
			CreatedAt:     goTime(row.CreatedAt),
			AccountID:     row.AccountID,
			AccountName:   accountName,
		})
	}
	return txs, nil
}

// ListByAccountWithDateRange returns paginated transactions for an account filtered by date range.
func (r *TransactionRepository) ListByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time, limit, offset int) ([]*domain.Transaction, error) {
	rows, err := r.q.ListTransactionsByAccountWithDateRange(ctx, sqlcgen.ListTransactionsByAccountWithDateRangeParams{
		UserID:    userID,
		AccountID: accountID,
		Limit:     int32(limit),
		Offset:    int32(offset),
		Column5:   pgOptionalTimestamptz(from),
		Column6:   pgOptionalTimestamptz(to),
	})
	if err != nil {
		return nil, err
	}
	txs := make([]*domain.Transaction, 0, len(rows))
	for _, row := range rows {
		var accountName string
		if row.AccountName.Valid {
			accountName = row.AccountName.String
		}
		txs = append(txs, &domain.Transaction{
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          row.Type,
			AmountCents:   row.AmountCents,
			CategoryID:    row.CategoryID,
			CategoryName:  row.CategoryName,
			CategoryIcon:  row.CategoryIcon,
			CategoryColor: row.CategoryColor,
			Note:          row.Note,
			CurrencyCode:  row.CurrencyCode,
			CreatedAt:     goTime(row.CreatedAt),
			AccountID:     row.AccountID,
			AccountName:   accountName,
		})
	}
	return txs, nil
}

// CountWithDateRange counts transactions for a user within an optional date range.
func (r *TransactionRepository) CountWithDateRange(ctx context.Context, userID int64, from, to *time.Time) (int64, error) {
	return r.q.CountUserTransactionsWithDateRange(ctx, sqlcgen.CountUserTransactionsWithDateRangeParams{
		UserID:  userID,
		Column2: pgOptionalTimestamptz(from),
		Column3: pgOptionalTimestamptz(to),
	})
}

// CountByAccountWithDateRange counts transactions for an account within an optional date range.
func (r *TransactionRepository) CountByAccountWithDateRange(ctx context.Context, userID, accountID int64, from, to *time.Time) (int64, error) {
	return r.q.CountUserTransactionsByAccountWithDateRange(ctx, sqlcgen.CountUserTransactionsByAccountWithDateRangeParams{
		UserID:    userID,
		AccountID: accountID,
		Column3:   pgOptionalTimestamptz(from),
		Column4:   pgOptionalTimestamptz(to),
	})
}

// GetBalanceByCurrencyAndAccount returns per-currency income/expense for a specific account.
func (r *TransactionRepository) GetBalanceByCurrencyAndAccount(ctx context.Context, userID, accountID int64) ([]domain.BalanceByCurrency, error) {
	rows, err := r.q.GetBalanceByCurrencyAndAccount(ctx, sqlcgen.GetBalanceByCurrencyAndAccountParams{
		UserID:    userID,
		AccountID: accountID,
	})
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
