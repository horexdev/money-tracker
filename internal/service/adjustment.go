package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// AdjustmentService handles balance correction entries.
// An adjustment affects the account balance but is excluded from
// transaction history and statistics.
type AdjustmentService struct {
	txRepo   TransactionStorer
	accounts AccountStorer
	catRepo  CategoryStorer
	log      *slog.Logger
}

// NewAdjustmentService constructs an AdjustmentService.
func NewAdjustmentService(txRepo TransactionStorer, accounts AccountStorer, catRepo CategoryStorer, log *slog.Logger) *AdjustmentService {
	return &AdjustmentService{
		txRepo:   txRepo,
		accounts: accounts,
		catRepo:  catRepo,
		log:      log,
	}
}

// Apply creates a balance-adjustment transaction for the given account.
// deltaCents is the signed amount to add to the balance:
//   - positive → income-type transaction (balance increases)
//   - negative → expense-type transaction (balance decreases)
//
// The resulting transaction has is_adjustment = true, so it is hidden
// from history lists and category statistics.
func (s *AdjustmentService) Apply(ctx context.Context, userID, accountID, deltaCents int64, note string) (*domain.Transaction, error) {
	if deltaCents == 0 {
		return nil, domain.ErrAdjustmentZeroAmount
	}

	acc, err := s.accounts.GetByID(ctx, accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	cat, err := s.catRepo.GetSystemCategoryByType(ctx, string(domain.CategoryTypeAdjustment))
	if err != nil {
		return nil, fmt.Errorf("get adjustment category: %w", err)
	}

	txType := domain.TransactionTypeIncome
	amountCents := deltaCents
	if deltaCents < 0 {
		txType = domain.TransactionTypeExpense
		amountCents = -deltaCents
	}

	tx, err := s.txRepo.CreateAdjustment(ctx, &domain.Transaction{
		UserID:       userID,
		Type:         txType,
		AmountCents:  amountCents,
		CategoryID:   cat.ID,
		Note:         note,
		CurrencyCode: acc.CurrencyCode,
		AccountID:    accountID,
		SnapshotDate: time.Now().UTC().Truncate(24 * time.Hour),
	})
	if err != nil {
		return nil, fmt.Errorf("create adjustment transaction: %w", err)
	}

	s.log.InfoContext(ctx, "balance adjustment applied",
		slog.Int64("user_id", userID),
		slog.Int64("account_id", accountID),
		slog.Int64("delta_cents", deltaCents),
		slog.String("currency", acc.CurrencyCode),
		slog.Int64("tx_id", tx.ID),
	)
	return tx, nil
}
