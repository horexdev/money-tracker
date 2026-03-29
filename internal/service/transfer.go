package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TransferService handles business logic for account transfers.
type TransferService struct {
	transfers TransferStorer
	accounts  AccountStorer
	goals     SavingsGoalStorer
	log       *slog.Logger
}

// NewTransferService constructs a TransferService.
func NewTransferService(transfers TransferStorer, accounts AccountStorer, goals SavingsGoalStorer, log *slog.Logger) *TransferService {
	return &TransferService{
		transfers: transfers,
		accounts:  accounts,
		goals:     goals,
		log:       log,
	}
}

// Execute moves funds from one account to another.
// It records the transfer and, if the destination account is linked to a savings
// goal, auto-increments the goal's current_cents.
func (s *TransferService) Execute(ctx context.Context, userID, fromAccountID, toAccountID, amountCents int64, exchangeRate float64, note string) (*domain.Transfer, error) {
	if fromAccountID == toAccountID {
		return nil, domain.ErrTransferSameAccount
	}
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	fromAcc, err := s.accounts.GetByID(ctx, fromAccountID, userID)
	if err != nil {
		return nil, fmt.Errorf("get from-account: %w", err)
	}
	toAcc, err := s.accounts.GetByID(ctx, toAccountID, userID)
	if err != nil {
		return nil, fmt.Errorf("get to-account: %w", err)
	}

	if exchangeRate <= 0 {
		exchangeRate = 1.0
	}

	t := &domain.Transfer{
		UserID:           userID,
		FromAccountID:    fromAccountID,
		ToAccountID:      toAccountID,
		FromAccountName:  fromAcc.Name,
		ToAccountName:    toAcc.Name,
		AmountCents:      amountCents,
		FromCurrencyCode: fromAcc.CurrencyCode,
		ToCurrencyCode:   toAcc.CurrencyCode,
		ExchangeRate:     exchangeRate,
		Note:             note,
	}

	created, err := s.transfers.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("create transfer record: %w", err)
	}

	// Best-effort: auto-increment linked savings goals on destination account.
	linkedGoals, err := s.goals.GetByAccountID(ctx, toAccountID)
	if err != nil {
		s.log.WarnContext(ctx, "failed to fetch linked goals for auto-increment",
			slog.Int64("to_account_id", toAccountID),
			slog.String("err", err.Error()),
		)
	}
	toAmountCents := int64(float64(amountCents) * exchangeRate)
	for _, g := range linkedGoals {
		if _, err := s.goals.Deposit(ctx, g.ID, userID, toAmountCents); err != nil {
			s.log.WarnContext(ctx, "failed to auto-increment goal on transfer",
				slog.Int64("goal_id", g.ID),
				slog.String("err", err.Error()),
			)
		}
	}

	s.log.InfoContext(ctx, "transfer executed",
		slog.Int64("user_id", userID),
		slog.Int64("transfer_id", created.ID),
		slog.Int64("from_account_id", fromAccountID),
		slog.Int64("to_account_id", toAccountID),
		slog.Int64("amount_cents", amountCents),
	)
	return created, nil
}

// List returns paginated transfers for a user.
func (s *TransferService) List(ctx context.Context, userID int64, limit, offset int) ([]*domain.Transfer, error) {
	transfers, err := s.transfers.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list transfers: %w", err)
	}
	return transfers, nil
}

// Count returns the total number of transfers for a user.
func (s *TransferService) Count(ctx context.Context, userID int64) (int64, error) {
	n, err := s.transfers.Count(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count transfers: %w", err)
	}
	return n, nil
}

// GetByID returns a single transfer.
func (s *TransferService) GetByID(ctx context.Context, id, userID int64) (*domain.Transfer, error) {
	t, err := s.transfers.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get transfer: %w", err)
	}
	return t, nil
}

// Delete removes a transfer record.
func (s *TransferService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.transfers.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete transfer: %w", err)
	}
	return nil
}

// ListByAccount returns all transfers involving a specific account.
func (s *TransferService) ListByAccount(ctx context.Context, userID, accountID int64) ([]*domain.Transfer, error) {
	transfers, err := s.transfers.ListByAccount(ctx, userID, accountID)
	if err != nil {
		return nil, fmt.Errorf("list transfers by account: %w", err)
	}
	return transfers, nil
}
