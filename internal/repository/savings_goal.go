package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// SavingsGoalRepository handles persistence of SavingsGoal entities.
type SavingsGoalRepository struct {
	q *sqlcgen.Queries
}

func NewSavingsGoalRepository(pool *pgxpool.Pool) *SavingsGoalRepository {
	return &SavingsGoalRepository{q: sqlcgen.New(pool)}
}

// Create inserts a new savings goal.
func (r *SavingsGoalRepository) Create(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	row, err := r.q.CreateSavingsGoal(ctx, sqlcgen.CreateSavingsGoalParams{
		UserID:       g.UserID,
		Name:         g.Name,
		TargetCents:  g.TargetCents,
		CurrencyCode: g.CurrencyCode,
		Deadline:     pgDate(g.Deadline),
	})
	if err != nil {
		return nil, err
	}
	return rowToGoal(row), nil
}

// GetByID returns a savings goal scoped to user.
func (r *SavingsGoalRepository) GetByID(ctx context.Context, id, userID int64) (*domain.SavingsGoal, error) {
	row, err := r.q.GetSavingsGoalByID(ctx, sqlcgen.GetSavingsGoalByIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	return rowToGoal(row), nil
}

// ListByUser returns all savings goals for a user.
func (r *SavingsGoalRepository) ListByUser(ctx context.Context, userID int64) ([]*domain.SavingsGoal, error) {
	rows, err := r.q.ListSavingsGoalsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	goals := make([]*domain.SavingsGoal, 0, len(rows))
	for _, row := range rows {
		goals = append(goals, rowToGoal(row))
	}
	return goals, nil
}

// Update modifies an existing savings goal.
func (r *SavingsGoalRepository) Update(ctx context.Context, g *domain.SavingsGoal) (*domain.SavingsGoal, error) {
	row, err := r.q.UpdateSavingsGoal(ctx, sqlcgen.UpdateSavingsGoalParams{
		ID:           g.ID,
		UserID:       g.UserID,
		Name:         g.Name,
		TargetCents:  g.TargetCents,
		CurrencyCode: g.CurrencyCode,
		Deadline:     pgDate(g.Deadline),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	return rowToGoal(row), nil
}

// Deposit adds funds to a savings goal.
func (r *SavingsGoalRepository) Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	row, err := r.q.DepositToGoal(ctx, sqlcgen.DepositToGoalParams{
		ID:           id,
		UserID:       userID,
		CurrentCents: amountCents,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGoalNotFound
		}
		return nil, err
	}
	return rowToGoal(row), nil
}

// Withdraw removes funds from a savings goal.
func (r *SavingsGoalRepository) Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	row, err := r.q.WithdrawFromGoal(ctx, sqlcgen.WithdrawFromGoalParams{
		ID:           id,
		UserID:       userID,
		CurrentCents: amountCents,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInsufficientGoalFunds
		}
		return nil, err
	}
	return rowToGoal(row), nil
}

// Delete removes a savings goal.
func (r *SavingsGoalRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.q.DeleteSavingsGoal(ctx, sqlcgen.DeleteSavingsGoalParams{ID: id, UserID: userID})
}

func rowToGoal(row sqlcgen.SavingsGoal) *domain.SavingsGoal {
	g := &domain.SavingsGoal{
		ID:           row.ID,
		UserID:       row.UserID,
		Name:         row.Name,
		TargetCents:  row.TargetCents,
		CurrentCents: row.CurrentCents,
		CurrencyCode: row.CurrencyCode,
		Deadline:     goDatePtr(row.Deadline),
		CreatedAt:    goTime(row.CreatedAt),
		UpdatedAt:    goTime(row.UpdatedAt),
	}
	return g
}
