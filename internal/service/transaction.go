package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TransactionService handles business logic for recording and querying transactions.
type TransactionService struct {
	txRepo  TransactionStorer
	catRepo CategoryStorer
	log     *slog.Logger
}

func NewTransactionService(
	txRepo TransactionStorer,
	catRepo CategoryStorer,
	log *slog.Logger,
) *TransactionService {
	return &TransactionService{txRepo: txRepo, catRepo: catRepo, log: log}
}

// AddExpense records a new expense transaction.
// exchangeRate is the rate from currencyCode to baseCurrency at creation time (1.0 if same currency).
// createdAt is optional; when nil the DB defaults to NOW().
func (s *TransactionService) AddExpense(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode, baseCurrency string, exchangeRate float64, accountID *int64, createdAt *time.Time) (*domain.Transaction, error) {
	return s.add(ctx, userID, domain.TransactionTypeExpense, amountCents, categoryID, note, currencyCode, baseCurrency, exchangeRate, accountID, createdAt)
}

// AddIncome records a new income transaction.
// exchangeRate is the rate from currencyCode to baseCurrency at creation time (1.0 if same currency).
// createdAt is optional; when nil the DB defaults to NOW().
func (s *TransactionService) AddIncome(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode, baseCurrency string, exchangeRate float64, accountID *int64, createdAt *time.Time) (*domain.Transaction, error) {
	return s.add(ctx, userID, domain.TransactionTypeIncome, amountCents, categoryID, note, currencyCode, baseCurrency, exchangeRate, accountID, createdAt)
}

func (s *TransactionService) add(ctx context.Context, userID int64, txType domain.TransactionType, amountCents, categoryID int64, note, currencyCode, baseCurrency string, exchangeRate float64, accountID *int64, createdAt *time.Time) (*domain.Transaction, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if currencyCode == "" {
		currencyCode = "USD"
	}
	if baseCurrency == "" {
		baseCurrency = currencyCode
	}
	if exchangeRate <= 0 {
		exchangeRate = 1.0
	}

	// Verify category exists and belongs to this user or is a system category.
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if !cat.IsSystem() && cat.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}

	var aid int64
	if accountID != nil {
		aid = *accountID
	}
	t := &domain.Transaction{
		UserID:                 userID,
		Type:                   txType,
		AmountCents:            amountCents,
		CategoryID:             categoryID,
		Note:                   note,
		CurrencyCode:           currencyCode,
		ExchangeRateSnapshot:   exchangeRate,
		BaseCurrencyAtCreation: baseCurrency,
		AccountID:              aid,
	}

	var tx *domain.Transaction
	if createdAt != nil {
		t.CreatedAt = *createdAt
		tx, err = s.txRepo.CreateWithDate(ctx, t)
	} else {
		tx, err = s.txRepo.Create(ctx, t)
	}
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	tx.CategoryName = cat.Name
	tx.CategoryEmoji = cat.Emoji
	tx.CategoryColor = cat.Color

	s.log.InfoContext(ctx, "transaction recorded",
		slog.Int64("user_id", userID),
		slog.String("type", string(txType)),
		slog.Int64("amount_cents", amountCents),
		slog.Int64("category_id", categoryID),
		slog.String("currency", currencyCode),
		slog.String("base_currency", baseCurrency),
		slog.Float64("exchange_rate_snapshot", exchangeRate),
	)
	return tx, nil
}

// UpdateTransaction modifies amount_cents, category, note, and created_at of an existing transaction.
func (s *TransactionService) UpdateTransaction(ctx context.Context, userID, id, amountCents, categoryID int64, note string, createdAt time.Time) (*domain.Transaction, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("get category %d: %w", categoryID, err)
	}
	if !cat.IsSystem() && cat.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}
	tx, err := s.txRepo.Update(ctx, &domain.Transaction{
		ID:          id,
		UserID:      userID,
		AmountCents: amountCents,
		CategoryID:  categoryID,
		Note:        note,
		CreatedAt:   createdAt,
	})
	if err != nil {
		return nil, fmt.Errorf("update transaction %d: %w", id, err)
	}
	tx.CategoryName = cat.Name
	tx.CategoryEmoji = cat.Emoji
	tx.CategoryColor = cat.Color
	s.log.InfoContext(ctx, "transaction updated",
		slog.Int64("user_id", userID),
		slog.Int64("transaction_id", id),
	)
	return tx, nil
}

// Delete removes a transaction by ID, ensuring it belongs to the given user.
func (s *TransactionService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.txRepo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete transaction %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "transaction deleted",
		slog.Int64("user_id", userID),
		slog.Int64("transaction_id", id),
	)
	return nil
}

// GetTotalInBaseCurrency returns net balance in base currency using historical exchange rate snapshots.
func (s *TransactionService) GetTotalInBaseCurrency(ctx context.Context, userID int64) (int64, error) {
	total, err := s.txRepo.GetTotalInBaseCurrency(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get total in base currency for user %d: %w", userID, err)
	}
	return total, nil
}

// GetBalanceByCurrency returns per-currency income/expense totals for a user.
func (s *TransactionService) GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error) {
	balances, err := s.txRepo.GetBalanceByCurrency(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get balance by currency for user %d: %w", userID, err)
	}
	return balances, nil
}

// GetBalance returns net balance (income - expense) for a user.
func (s *TransactionService) GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error) {
	income, expense, err := s.txRepo.GetBalance(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("get balance for user %d: %w", userID, err)
	}
	return income, expense, nil
}

// ListRecent returns the n most recent transactions for a user.
func (s *TransactionService) ListRecent(ctx context.Context, userID int64, n int) ([]*domain.Transaction, error) {
	txs, err := s.txRepo.List(ctx, userID, n, 0)
	if err != nil {
		return nil, fmt.Errorf("list transactions for user %d: %w", userID, err)
	}
	return txs, nil
}

// ListPaged returns a page of transactions and the total page count.
func (s *TransactionService) ListPaged(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Transaction, int, error) {
	total, err := s.txRepo.Count(ctx, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions for user %d: %w", userID, err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * pageSize
	txs, err := s.txRepo.List(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions for user %d: %w", userID, err)
	}
	return txs, totalPages, nil
}

// ListPagedByAccount returns a page of transactions for a specific account.
func (s *TransactionService) ListPagedByAccount(ctx context.Context, userID, accountID int64, page, pageSize int) ([]*domain.Transaction, int, error) {
	total, err := s.txRepo.CountByAccount(ctx, userID, accountID)
	if err != nil {
		return nil, 0, fmt.Errorf("count transactions for account %d: %w", accountID, err)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * pageSize
	txs, err := s.txRepo.ListByAccount(ctx, userID, accountID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions for account %d: %w", accountID, err)
	}
	return txs, totalPages, nil
}

// GetBalanceByCurrencyAndAccount returns per-currency balance for a specific account.
func (s *TransactionService) GetBalanceByCurrencyAndAccount(ctx context.Context, userID, accountID int64) ([]domain.BalanceByCurrency, error) {
	return s.txRepo.GetBalanceByCurrencyAndAccount(ctx, userID, accountID)
}

// ListCategories returns all categories available to a user.
func (s *TransactionService) ListCategories(ctx context.Context, userID int64) ([]*domain.Category, error) {
	cats, err := s.catRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list categories for user %d: %w", userID, err)
	}
	return cats, nil
}
