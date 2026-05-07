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

func newAdminRepoFixtures(t *testing.T) (*repository.AdminRepository, func() int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	seq := time.Now().UnixNano()
	next := func() int64 {
		seq++
		return seq
	}
	return repository.NewAdminRepository(pool), next
}

func TestAdminRepository_CountUsers_EmptyAndAfterSeed(t *testing.T) {
	repo, next := newAdminRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	ctx := context.Background()

	n, err := repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)

	testutil.SeedUser(t, pool, next())
	testutil.SeedUser(t, pool, next())
	testutil.SeedUser(t, pool, next())

	n, err = repo.CountUsers(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3), n)
}

func TestAdminRepository_ListUsers_Pagination(t *testing.T) {
	repo, next := newAdminRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		testutil.SeedUser(t, pool, next())
	}

	page1, err := repo.ListUsers(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	page2, err := repo.ListUsers(ctx, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 2)

	page3, err := repo.ListUsers(ctx, 2, 4)
	require.NoError(t, err)
	assert.Len(t, page3, 1)
}

func TestAdminRepository_ListAllUserIDs(t *testing.T) {
	repo, next := newAdminRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	ctx := context.Background()

	id1 := next()
	id2 := next()
	testutil.SeedUser(t, pool, id1)
	testutil.SeedUser(t, pool, id2)

	ids, err := repo.ListAllUserIDs(ctx)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
	assert.Contains(t, ids, id1)
	assert.Contains(t, ids, id2)
}

func TestAdminRepository_CountNewUsers_RespectsRange(t *testing.T) {
	repo, next := newAdminRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	ctx := context.Background()

	testutil.SeedUser(t, pool, next())
	testutil.SeedUser(t, pool, next())

	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	n, err := repo.CountNewUsers(ctx, from, to)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)

	// Range that does not include the seeded users
	pastFrom := time.Now().Add(-48 * time.Hour)
	pastTo := time.Now().Add(-24 * time.Hour)
	n, err = repo.CountNewUsers(ctx, pastFrom, pastTo)
	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
}
