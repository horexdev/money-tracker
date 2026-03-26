package service

import (
	"context"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository"
)

// TransactionService handles business logic for recording and querying transactions.
type TransactionService struct {
	txRepo  *repository.TransactionRepository
	catRepo *repository.CategoryRepository
	log     *slog.Logger
}

func NewTransactionService(
	txRepo *repository.TransactionRepository,
	catRepo *repository.CategoryRepository,
	log *slog.Logger,
) *TransactionService {
	return &TransactionService{txRepo: txRepo, catRepo: catRepo, log: log}
}

// AddExpense records a new expense transaction.
func (s *TransactionService) AddExpense(ctx context.Context, userID, amountCents, categoryID int64, note string) (*domain.Transaction, error) {
	return s.add(ctx, userID, domain.TransactionTypeExpense, amountCents, categoryID, note)
}

// AddIncome records a new income transaction.
func (s *TransactionService) AddIncome(ctx context.Context, userID, amountCents, categoryID int64, note string) (*domain.Transaction, error) {
	return s.add(ctx, userID, domain.TransactionTypeIncome, amountCents, categoryID, note)
}

func (s *TransactionService) add(ctx context.Context, userID int64, txType domain.TransactionType, amountCents, categoryID int64, note string) (*domain.Transaction, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	// Verify category exists and belongs to this user or is a system category.
	cat, err := s.catRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if !cat.IsSystem() && cat.UserID != userID {
		return nil, domain.ErrCategoryNotFound
	}

	tx, err := s.txRepo.Create(ctx, &domain.Transaction{
		UserID:      userID,
		Type:        txType,
		AmountCents: amountCents,
		CategoryID:  categoryID,
		Note:        note,
	})
	if err != nil {
		s.log.ErrorContext(ctx, "failed to create transaction",
			slog.Int64("user_id", userID),
			slog.String("type", string(txType)),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	tx.CategoryName = cat.Name
	tx.CategoryEmoji = cat.Emoji

	s.log.InfoContext(ctx, "transaction recorded",
		slog.Int64("user_id", userID),
		slog.String("type", string(txType)),
		slog.Int64("amount_cents", amountCents),
		slog.Int64("category_id", categoryID),
	)
	return tx, nil
}

// GetBalance returns net balance (income - expense) for a user.
func (s *TransactionService) GetBalance(ctx context.Context, userID int64) (incomeCents, expenseCents int64, err error) {
	return s.txRepo.GetBalance(ctx, userID)
}

// ListRecent returns the n most recent transactions for a user.
func (s *TransactionService) ListRecent(ctx context.Context, userID int64, n int) ([]*domain.Transaction, error) {
	return s.txRepo.List(ctx, userID, n, 0)
}
