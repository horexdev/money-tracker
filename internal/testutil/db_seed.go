//go:build integration

package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// SeedUser inserts a minimal users row with the given Telegram ID. The other
// columns rely on the schema's defaults. Returns the user ID for chaining.
func SeedUser(t *testing.T, pool *pgxpool.Pool, id int64) int64 {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO users (id, username, first_name)
		VALUES ($1, $2, $3)
	`, id, "test_user", "Test")
	require.NoError(t, err, "seed user")
	return id
}

// SeedAccount inserts an account for the given user and returns its ID.
// The account is non-default and included in the user's total.
func SeedAccount(t *testing.T, pool *pgxpool.Pool, userID int64, currency string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO accounts (user_id, name, currency_code, is_default, include_in_total)
		VALUES ($1, 'Test Account', $2, false, true)
		RETURNING id
	`, userID, currency).Scan(&id)
	require.NoError(t, err, "seed account")
	return id
}

// SeedCategory inserts a personal category for the user and returns its ID.
// catType must be one of: expense, income, both, transfer, adjustment, savings.
func SeedCategory(t *testing.T, pool *pgxpool.Pool, userID int64, name, catType string) int64 {
	t.Helper()
	var id int64
	err := pool.QueryRow(context.Background(), `
		INSERT INTO categories (user_id, name, icon, type, color)
		VALUES ($1, $2, 'package', $3, '#6366f1')
		RETURNING id
	`, userID, name, catType).Scan(&id)
	require.NoError(t, err, "seed category")
	return id
}
