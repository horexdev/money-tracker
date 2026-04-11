package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// SavingsGoalService handles business logic for savings goals.
type SavingsGoalService struct {
	repo        SavingsGoalStorer
	txRepo      TransactionStorer
	catRepo     CategoryStorer
	accountRepo AccountStorer
	log         *slog.Logger
}

func NewSavingsGoalService(repo SavingsGoalStorer, txRepo TransactionStorer, catRepo CategoryStorer, accountRepo AccountStorer, log *slog.Logger) *SavingsGoalService {
	return &SavingsGoalService{repo: repo, txRepo: txRepo, catRepo: catRepo, accountRepo: accountRepo, log: log}
}

// Create adds a new savings goal.
func (s *SavingsGoalService) Create(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	if g.TargetCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.Create(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("create savings goal: %w", err)
	}

	s.log.InfoContext(ctx, "savings goal created",
		slog.Int64("user_id", g.UserID),
		slog.Int64("goal_id", goal.ID),
		slog.String("name", g.Name),
	)
	return goal, nil
}

// GetByID returns a savings goal scoped to user.
func (s *SavingsGoalService) GetByID(ctx context.Context, id, userID int64) (*domain.SavingsGoal, error) {
	return s.repo.GetByID(ctx, id, userID)
}

// ListByUser returns all savings goals for a user.
func (s *SavingsGoalService) ListByUser(ctx context.Context, userID int64) ([]*domain.SavingsGoal, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Update modifies an existing savings goal.
func (s *SavingsGoalService) Update(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	if g.TargetCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.Update(ctx, g)
	if err != nil {
		return nil, fmt.Errorf("update savings goal %d: %w", g.ID, err)
	}
	return goal, nil
}

// Deposit adds funds to a savings goal.
// If the goal has a linked account, an expense transaction is created on that account.
func (s *SavingsGoalService) Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get goal %d: %w", id, err)
	}

	if goal.AccountID != nil {
		if err := s.createLinkedTransaction(ctx, userID, *goal.AccountID, amountCents, domain.TransactionTypeExpense, goal.Name); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.Deposit(ctx, id, userID, amountCents)
	if err != nil {
		return nil, fmt.Errorf("deposit to goal %d: %w", id, err)
	}

	s.log.InfoContext(ctx, "deposit to savings goal",
		slog.Int64("goal_id", id),
		slog.Int64("amount_cents", amountCents),
	)
	return updated, nil
}

// Withdraw removes funds from a savings goal.
// If the goal has a linked account, an income transaction is created on that account.
func (s *SavingsGoalService) Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get goal %d: %w", id, err)
	}

	if goal.AccountID != nil {
		if err := s.createLinkedTransaction(ctx, userID, *goal.AccountID, amountCents, domain.TransactionTypeIncome, goal.Name); err != nil {
			return nil, err
		}
	}

	updated, err := s.repo.Withdraw(ctx, id, userID, amountCents)
	if err != nil {
		return nil, fmt.Errorf("withdraw from goal %d: %w", id, err)
	}

	s.log.InfoContext(ctx, "withdraw from savings goal",
		slog.Int64("goal_id", id),
		slog.Int64("amount_cents", amountCents),
	)
	return updated, nil
}

// createLinkedTransaction creates a real expense/income transaction on the linked account.
// Currency is read from the account itself, not from transactions.
func (s *SavingsGoalService) createLinkedTransaction(ctx context.Context, userID, accountID, amountCents int64, txType domain.TransactionType, note string) error {
	cat, err := s.catRepo.GetBySavingsType(ctx, userID)
	if err != nil {
		return fmt.Errorf("get savings category: %w", err)
	}

	acc, err := s.accountRepo.GetByID(ctx, accountID, userID)
	if err != nil {
		return fmt.Errorf("get linked account: %w", err)
	}

	tx := &domain.Transaction{
		UserID:       userID,
		AmountCents:  amountCents,
		CategoryID:   cat.ID,
		Type:         txType,
		Note:         note,
		CurrencyCode: acc.CurrencyCode,
		AccountID:    accountID,
		SnapshotDate: time.Now().UTC().Truncate(24 * time.Hour),
	}
	if _, err := s.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("create linked transaction for goal: %w", err)
	}
	return nil
}

// Delete removes a savings goal.
func (s *SavingsGoalService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete savings goal %d: %w", id, err)
	}
	return nil
}

// ListHistory returns the deposit/withdraw history for a savings goal.
func (s *SavingsGoalService) ListHistory(ctx context.Context, goalID, userID int64) ([]*domain.GoalTransaction, error) {
	history, err := s.repo.ListHistory(ctx, goalID, userID)
	if err != nil {
		return nil, fmt.Errorf("list history for goal %d: %w", goalID, err)
	}
	return history, nil
}
