package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/horexdev/money-tracker/internal/domain"
)

// SavingsGoalService handles business logic for savings goals.
type SavingsGoalService struct {
	repo SavingsGoalStorer
	log  *slog.Logger
}

func NewSavingsGoalService(repo SavingsGoalStorer, log *slog.Logger) *SavingsGoalService {
	return &SavingsGoalService{repo: repo, log: log}
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
func (s *SavingsGoalService) Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.Deposit(ctx, id, userID, amountCents)
	if err != nil {
		return nil, fmt.Errorf("deposit to goal %d: %w", id, err)
	}

	s.log.InfoContext(ctx, "deposit to savings goal",
		slog.Int64("goal_id", id),
		slog.Int64("amount_cents", amountCents),
	)
	return goal, nil
}

// Withdraw removes funds from a savings goal.
func (s *SavingsGoalService) Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	if amountCents <= 0 {
		return nil, domain.ErrInvalidAmount
	}

	goal, err := s.repo.Withdraw(ctx, id, userID, amountCents)
	if err != nil {
		return nil, fmt.Errorf("withdraw from goal %d: %w", id, err)
	}

	s.log.InfoContext(ctx, "withdraw from savings goal",
		slog.Int64("goal_id", id),
		slog.Int64("amount_cents", amountCents),
	)
	return goal, nil
}

// Delete removes a savings goal.
func (s *SavingsGoalService) Delete(ctx context.Context, id, userID int64) error {
	if err := s.repo.Delete(ctx, id, userID); err != nil {
		return fmt.Errorf("delete savings goal %d: %w", id, err)
	}
	return nil
}
