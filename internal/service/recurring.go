package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// RecurringService handles business logic for recurring transactions.
type RecurringService struct {
	repo    RecurringStorer
	txRepo  TransactionStorer
	accRepo AccountStorer
	log     *slog.Logger
}

func NewRecurringService(repo RecurringStorer, txRepo TransactionStorer, accRepo AccountStorer, log *slog.Logger) *RecurringService {
	return &RecurringService{repo: repo, txRepo: txRepo, accRepo: accRepo, log: log}
}

// Create adds a new recurring transaction.
func (s *RecurringService) Create(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	if rt.AmountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if !validFrequency(rt.Frequency) {
		return nil, domain.ErrInvalidFrequency
	}
	if rt.AccountID == 0 {
		return nil, domain.ErrAccountNotFound
	}

	if rt.NextRunAt.IsZero() {
		rt.NextRunAt = rt.NextRunAfter(time.Now())
	}

	result, err := s.repo.Create(ctx, rt)
	if err != nil {
		return nil, fmt.Errorf("create recurring transaction: %w", err)
	}

	s.log.InfoContext(ctx, "recurring transaction created",
		slog.Int64("user_id", rt.UserID),
		slog.Int64("recurring_id", result.ID),
		slog.String("frequency", string(rt.Frequency)),
	)
	return result, nil
}

// GetByID returns a recurring transaction scoped to user.
func (s *RecurringService) GetByID(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListByUser returns all recurring transactions for a user.
func (s *RecurringService) ListByUser(ctx context.Context, userID int64) ([]*domain.RecurringTransaction, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Update modifies an existing recurring transaction.
func (s *RecurringService) Update(ctx context.Context, rt *domain.RecurringTransaction) (*domain.RecurringTransaction, error) {
	if rt.AmountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if !validFrequency(rt.Frequency) {
		return nil, domain.ErrInvalidFrequency
	}

	result, err := s.repo.Update(ctx, rt)
	if err != nil {
		return nil, fmt.Errorf("update recurring transaction %d: %w", rt.ID, err)
	}
	return result, nil
}

// ToggleActive flips the is_active flag.
func (s *RecurringService) ToggleActive(ctx context.Context, id, userID int64) (*domain.RecurringTransaction, error) {
	result, err := s.repo.ToggleActive(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("toggle recurring %d: %w", id, err)
	}
	return result, nil
}

// Delete removes a recurring transaction.
func (s *RecurringService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete recurring %d: %w", id, err)
	}
	return nil
}

// ProcessDue finds all due recurring transactions, creates actual transactions
// with the account's currency, and advances next_run_at.
func (s *RecurringService) ProcessDue(ctx context.Context) (int, error) {
	now := time.Now()
	due, err := s.repo.GetDue(ctx, now)
	if err != nil {
		return 0, fmt.Errorf("get due recurring: %w", err)
	}

	var processed int
	for _, rt := range due {
		// Look up the account to get its currency.
		acc, err := s.accRepo.GetByID(ctx, rt.AccountID, rt.UserID)
		if err != nil {
			s.log.ErrorContext(ctx, "recurring: account not found, skipping",
				slog.Int64("recurring_id", rt.ID),
				slog.Int64("account_id", rt.AccountID),
				slog.String("error", err.Error()),
			)
			continue
		}

		_, err = s.txRepo.Create(ctx, &domain.Transaction{
			UserID:       rt.UserID,
			Type:         rt.Type,
			AmountCents:  rt.AmountCents,
			CategoryID:   rt.CategoryID,
			Note:         rt.Note,
			CurrencyCode: acc.CurrencyCode,
			AccountID:    rt.AccountID,
			SnapshotDate: now.UTC().Truncate(24 * time.Hour),
		})
		if err != nil {
			s.log.ErrorContext(ctx, "failed to create transaction from recurring",
				slog.Int64("recurring_id", rt.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		nextRun := rt.NextRunAfter(now)
		if err := s.repo.UpdateNextRun(ctx, rt.ID, nextRun); err != nil {
			s.log.ErrorContext(ctx, "failed to update next_run_at",
				slog.Int64("recurring_id", rt.ID),
				slog.String("error", err.Error()),
			)
			continue
		}

		processed++
		s.log.InfoContext(ctx, "recurring transaction processed",
			slog.Int64("recurring_id", rt.ID),
			slog.Int64("user_id", rt.UserID),
			slog.Int64("account_id", rt.AccountID),
		)
	}
	return processed, nil
}

func validFrequency(f domain.Frequency) bool {
	switch f {
	case domain.FrequencyDaily, domain.FrequencyWeekly, domain.FrequencyMonthly, domain.FrequencyYearly:
		return true
	default:
		return false
	}
}
