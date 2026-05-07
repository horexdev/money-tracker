package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TransactionTemplateService handles business logic for user-defined transaction templates.
// Apply reuses TransactionService.AddExpense/AddIncome to inherit validation,
// snapshot_date computation, and exchange-rate lookup.
type TransactionTemplateService struct {
	repo    TransactionTemplateStorer
	txAdder TransactionAdder
	accRepo AccountStorer
	log     *slog.Logger
}

func NewTransactionTemplateService(
	repo TransactionTemplateStorer,
	txAdder TransactionAdder,
	accRepo AccountStorer,
	log *slog.Logger,
) *TransactionTemplateService {
	return &TransactionTemplateService{repo: repo, txAdder: txAdder, accRepo: accRepo, log: log}
}

// Create stores a new template. currency_code is inherited from the account when empty.
func (s *TransactionTemplateService) Create(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	if err := s.validate(t); err != nil {
		return nil, err
	}
	if t.CurrencyCode == "" {
		acc, err := s.accRepo.GetByID(ctx, t.AccountID, t.UserID)
		if err != nil {
			return nil, fmt.Errorf("get account %d: %w", t.AccountID, err)
		}
		t.CurrencyCode = acc.CurrencyCode
	}
	created, err := s.repo.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	s.log.InfoContext(ctx, "transaction template created",
		slog.Int64("user_id", t.UserID),
		slog.Int64("template_id", created.ID),
	)
	return created, nil
}

// GetByID returns a single template scoped to user.
func (s *TransactionTemplateService) GetByID(ctx context.Context, id, userID int64) (*domain.TransactionTemplate, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListByUser returns all templates for the user, ordered by sort_order.
func (s *TransactionTemplateService) ListByUser(ctx context.Context, userID int64) ([]*domain.TransactionTemplate, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Update modifies a template. currency_code is inherited from the account when empty.
func (s *TransactionTemplateService) Update(ctx context.Context, t *domain.TransactionTemplate) (*domain.TransactionTemplate, error) {
	if err := s.validate(t); err != nil {
		return nil, err
	}
	if t.CurrencyCode == "" {
		acc, err := s.accRepo.GetByID(ctx, t.AccountID, t.UserID)
		if err != nil {
			return nil, fmt.Errorf("get account %d: %w", t.AccountID, err)
		}
		t.CurrencyCode = acc.CurrencyCode
	}
	updated, err := s.repo.Update(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}
	s.log.InfoContext(ctx, "transaction template updated",
		slog.Int64("user_id", t.UserID),
		slog.Int64("template_id", t.ID),
	)
	return updated, nil
}

// Delete removes a template owned by the user.
func (s *TransactionTemplateService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete template %d: %w", id, err)
	}
	s.log.InfoContext(ctx, "transaction template deleted",
		slog.Int64("user_id", userID),
		slog.Int64("template_id", id),
	)
	return nil
}

// Reorder applies a new sort_order to the user's templates.
func (s *TransactionTemplateService) Reorder(ctx context.Context, userID int64, orderedIDs []int64) error {
	if err := s.repo.Reorder(ctx, userID, orderedIDs); err != nil {
		return fmt.Errorf("reorder templates for user %d: %w", userID, err)
	}
	return nil
}

// Apply creates a transaction from the given template. When overrideAmountCents is non-nil
// (variable-amount template), it overrides the template's stored amount.
func (s *TransactionTemplateService) Apply(ctx context.Context, id, userID int64, overrideAmountCents *int64) (*domain.Transaction, error) {
	tpl, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	amount := tpl.AmountCents
	if overrideAmountCents != nil {
		amount = *overrideAmountCents
	}
	if amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	switch tpl.Type {
	case domain.TransactionTypeExpense:
		return s.txAdder.AddExpense(ctx, userID, amount, tpl.CategoryID, tpl.Note, tpl.CurrencyCode, tpl.AccountID, nil)
	case domain.TransactionTypeIncome:
		return s.txAdder.AddIncome(ctx, userID, amount, tpl.CategoryID, tpl.Note, tpl.CurrencyCode, tpl.AccountID, nil)
	default:
		return nil, fmt.Errorf("unknown transaction type %q", tpl.Type)
	}
}

func (s *TransactionTemplateService) validate(t *domain.TransactionTemplate) error {
	if t.AmountCents <= 0 {
		return domain.ErrInvalidAmount
	}
	if t.Type != domain.TransactionTypeExpense && t.Type != domain.TransactionTypeIncome {
		return errors.New("invalid transaction type")
	}
	return nil
}
