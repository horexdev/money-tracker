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
	pool *pgxpool.Pool
	q    *sqlcgen.Queries
}

func NewSavingsGoalRepository(pool *pgxpool.Pool) *SavingsGoalRepository {
	return &SavingsGoalRepository{pool: pool, q: sqlcgen.New(pool)}
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

// Deposit adds funds to a savings goal and records a goal_transaction entry.
func (r *SavingsGoalRepository) Deposit(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.q.WithTx(tx)

	row, err := q.DepositToGoal(ctx, sqlcgen.DepositToGoalParams{
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

	if err := q.InsertGoalTransaction(ctx, sqlcgen.InsertGoalTransactionParams{
		GoalID:      id,
		UserID:      userID,
		Type:        "deposit",
		AmountCents: amountCents,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return rowToGoal(row), nil
}

// Withdraw removes funds from a savings goal and records a goal_transaction entry.
func (r *SavingsGoalRepository) Withdraw(ctx context.Context, id, userID, amountCents int64) (*domain.SavingsGoal, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := r.q.WithTx(tx)

	row, err := q.WithdrawFromGoal(ctx, sqlcgen.WithdrawFromGoalParams{
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

	if err := q.InsertGoalTransaction(ctx, sqlcgen.InsertGoalTransactionParams{
		GoalID:      id,
		UserID:      userID,
		Type:        "withdraw",
		AmountCents: amountCents,
	}); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return rowToGoal(row), nil
}

// Delete removes a savings goal.
func (r *SavingsGoalRepository) Delete(ctx context.Context, id, userID int64) error {
	return r.q.DeleteSavingsGoal(ctx, sqlcgen.DeleteSavingsGoalParams{ID: id, UserID: userID})
}

// ListHistory returns the deposit/withdraw history for a savings goal.
func (r *SavingsGoalRepository) ListHistory(ctx context.Context, goalID, userID int64) ([]*domain.GoalTransaction, error) {
	rows, err := r.q.ListGoalTransactions(ctx, sqlcgen.ListGoalTransactionsParams{
		GoalID: goalID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]*domain.GoalTransaction, 0, len(rows))
	for _, row := range rows {
		result = append(result, &domain.GoalTransaction{
			ID:          row.ID,
			GoalID:      row.GoalID,
			UserID:      row.UserID,
			Type:        row.Type,
			AmountCents: row.AmountCents,
			CreatedAt:   goTime(row.CreatedAt),
		})
	}
	return result, nil
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
