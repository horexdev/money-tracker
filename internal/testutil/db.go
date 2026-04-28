//go:build integration

package testutil

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// OpenTestPool opens a pgx pool to the integration-test database.
// It reads TEST_DATABASE_URL or falls back to DATABASE_URL — the same env
// variable the CI integration-tests job exports. If neither is set, the
// test is skipped so local `make test-unit` runs stay green.
func OpenTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL or DATABASE_URL not set; skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	require.NoError(t, err, "open pgx pool")
	t.Cleanup(pool.Close)

	require.NoError(t, pool.Ping(context.Background()), "ping db")
	return pool
}

// CleanupTables removes all user-scoped data, leaving system rows
// (e.g. seeded categories with user_id IS NULL) intact. Call it via
// t.Cleanup at the start of each test, and once more at the top to
// recover from a previously aborted run.
func CleanupTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	stmts := []string{
		`DELETE FROM goal_transactions`,
		`DELETE FROM savings_goals`,
		`DELETE FROM budgets`,
		`DELETE FROM recurring_transactions`,
		`DELETE FROM transfers`,
		`DELETE FROM transactions`,
		`DELETE FROM accounts`,
		`DELETE FROM exchange_rate_snapshots`,
		`DELETE FROM categories WHERE user_id IS NOT NULL`,
		`DELETE FROM users`,
	}
	for _, sql := range stmts {
		_, err := pool.Exec(ctx, sql)
		require.NoErrorf(t, err, "cleanup: %s", sql)
	}
}
