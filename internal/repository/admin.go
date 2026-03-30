package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// AdminRepository provides read-only queries for the admin panel.
type AdminRepository struct {
	q *sqlcgen.Queries
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{q: sqlcgen.New(pool)}
}

// ListUsers returns a paginated list of all users ordered by signup date descending.
func (r *AdminRepository) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.q.ListAllUsers(ctx, sqlcgen.ListAllUsersParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, rowToUser(row))
	}
	return users, nil
}

// CountUsers returns the total number of registered users.
func (r *AdminRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.q.CountAllUsers(ctx)
}

// CountNewUsers returns the number of users who signed up in [from, to).
func (r *AdminRepository) CountNewUsers(ctx context.Context, from, to time.Time) (int64, error) {
	return r.q.CountNewUsers(ctx, sqlcgen.CountNewUsersParams{
		FromTs: pgTimestamptz(from),
		ToTs:   pgTimestamptz(to),
	})
}

// CountActiveUsers returns the number of distinct users with transactions in [from, to).
func (r *AdminRepository) CountActiveUsers(ctx context.Context, from, to time.Time) (int64, error) {
	return r.q.CountActiveUsersInPeriod(ctx, sqlcgen.CountActiveUsersInPeriodParams{
		FromTs: pgTimestamptz(from),
		ToTs:   pgTimestamptz(to),
	})
}

// CountRetainedUsers returns the number of users who signed up in [signupFrom, signupTo)
// and also have at least one transaction in [activeFrom, activeTo).
func (r *AdminRepository) CountRetainedUsers(ctx context.Context, signupFrom, signupTo, activeFrom, activeTo time.Time) (int64, error) {
	return r.q.CountRetainedUsers(ctx, sqlcgen.CountRetainedUsersParams{
		SignupFrom: pgTimestamptz(signupFrom),
		SignupTo:   pgTimestamptz(signupTo),
		ActiveFrom: pgTimestamptz(activeFrom),
		ActiveTo:   pgTimestamptz(activeTo),
	})
}
