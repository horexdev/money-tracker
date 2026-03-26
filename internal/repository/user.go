package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// UserRepository handles persistence of User entities.
type UserRepository struct {
	q *sqlcgen.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: sqlcgen.New(pool)}
}

// Upsert creates a new user or updates username/name fields if already exists.
func (r *UserRepository) Upsert(ctx context.Context, u *domain.User) (*domain.User, error) {
	row, err := r.q.UpsertUser(ctx, sqlcgen.UpsertUserParams{
		ID:           u.ID,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		CurrencyCode: u.CurrencyCode,
	})
	if err != nil {
		return nil, err
	}
	return rowToUser(row), nil
}

// GetByID returns the user with the given Telegram ID.
// Returns domain.ErrUserNotFound if no such user exists.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return rowToUser(row), nil
}

func rowToUser(row sqlcgen.User) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Username:     row.Username,
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		CurrencyCode: row.CurrencyCode,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
