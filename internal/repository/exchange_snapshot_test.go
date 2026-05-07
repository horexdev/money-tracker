//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/testutil"
)

func newSnapshotRepoFixtures(t *testing.T) *repository.ExchangeSnapshotRepository {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })
	return repository.NewExchangeSnapshotRepository(pool)
}

func TestExchangeSnapshotRepository_UpsertAndGetRate(t *testing.T) {
	repo := newSnapshotRepoFixtures(t)
	ctx := context.Background()
	date := time.Now().UTC().Truncate(24 * time.Hour)

	require.NoError(t, repo.Upsert(ctx, date, "USD", "EUR", 0.92))

	rate, err := repo.GetRate(ctx, date, "USD", "EUR")
	require.NoError(t, err)
	assert.InDelta(t, 0.92, rate, 0.0001)
}

func TestExchangeSnapshotRepository_Upsert_OverridesExistingRate(t *testing.T) {
	repo := newSnapshotRepoFixtures(t)
	ctx := context.Background()
	date := time.Now().UTC().Truncate(24 * time.Hour)

	require.NoError(t, repo.Upsert(ctx, date, "USD", "EUR", 0.9))
	require.NoError(t, repo.Upsert(ctx, date, "USD", "EUR", 0.95))

	rate, err := repo.GetRate(ctx, date, "USD", "EUR")
	require.NoError(t, err)
	assert.InDelta(t, 0.95, rate, 0.0001)
}

func TestExchangeSnapshotRepository_GetRateOrLatest_FallsBackToPriorDate(t *testing.T) {
	repo := newSnapshotRepoFixtures(t)
	ctx := context.Background()
	yesterday := time.Now().UTC().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	today := yesterday.Add(24 * time.Hour)

	require.NoError(t, repo.Upsert(ctx, yesterday, "USD", "EUR", 0.92))

	// today not yet seeded; GetRateOrLatest should fall back to yesterday's rate.
	rate, err := repo.GetRateOrLatest(ctx, today, "USD", "EUR")
	require.NoError(t, err)
	assert.InDelta(t, 0.92, rate, 0.0001)
}

func TestExchangeSnapshotRepository_ListDistinctBaseCurrencies(t *testing.T) {
	// ListDistinctBaseCurrencies returns distinct currencies from the accounts
	// table (used by the daily cron to know which bases to fetch rates for),
	// not from exchange_rate_snapshots — seed accounts to verify.
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })
	repo := repository.NewExchangeSnapshotRepository(pool)

	userID := testutil.SeedUser(t, pool, 7777001)
	testutil.SeedAccount(t, pool, userID, "USD")
	testutil.SeedAccount(t, pool, userID, "EUR")
	testutil.SeedAccount(t, pool, userID, "USD") // duplicate currency, should be deduped

	bases, err := repo.ListDistinctBaseCurrencies(context.Background())
	require.NoError(t, err)
	assert.Contains(t, bases, "USD")
	assert.Contains(t, bases, "EUR")
}

func TestExchangeSnapshotRepository_GetLatestSnapshotDate(t *testing.T) {
	repo := newSnapshotRepoFixtures(t)
	ctx := context.Background()
	d1 := time.Now().UTC().Add(-48 * time.Hour).Truncate(24 * time.Hour)
	d2 := time.Now().UTC().Truncate(24 * time.Hour)

	require.NoError(t, repo.Upsert(ctx, d1, "USD", "EUR", 0.9))
	require.NoError(t, repo.Upsert(ctx, d2, "USD", "EUR", 0.95))

	latest, err := repo.GetLatestSnapshotDate(ctx)
	require.NoError(t, err)
	assert.Equal(t, d2.Year(), latest.Year())
	assert.Equal(t, d2.YearDay(), latest.YearDay())
}
