//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/testutil"
)

func newBudgetRepoFixtures(t *testing.T) (*repository.BudgetRepository, int64, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	categoryID := testutil.SeedCategory(t, pool, userID, "Food", "expense")
	return repository.NewBudgetRepository(pool), userID, categoryID
}

func TestBudgetRepository_Create_PersistsAttributes(t *testing.T) {
	repo, userID, categoryID := newBudgetRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Budget{
		UserID:               userID,
		CategoryID:           categoryID,
		LimitCents:           50000,
		Period:               domain.BudgetPeriodMonthly,
		CurrencyCode:         "USD",
		NotifyAtPercent:      80,
		NotificationsEnabled: true,
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, int64(50000), created.LimitCents)
	assert.Equal(t, domain.BudgetPeriodMonthly, created.Period)
	assert.Equal(t, 80, created.NotifyAtPercent)
	assert.True(t, created.NotificationsEnabled)
}

func TestBudgetRepository_GetByID_NotFound(t *testing.T) {
	repo, userID, _ := newBudgetRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99_999, userID)
	assert.ErrorIs(t, err, domain.ErrBudgetNotFound)
}

func TestBudgetRepository_ListByUser_OnlyOwnBudgets(t *testing.T) {
	repo, user1, cat1 := newBudgetRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	user2 := testutil.SeedUser(t, pool, time.Now().UnixNano()+1)
	cat2 := testutil.SeedCategory(t, pool, user2, "Travel", "expense")

	ctx := context.Background()
	_, err := repo.Create(ctx, &domain.Budget{
		UserID: user1, CategoryID: cat1, LimitCents: 10000,
		Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Budget{
		UserID: user2, CategoryID: cat2, LimitCents: 20000,
		Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	list, err := repo.ListByUser(ctx, user1)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, user1, list[0].UserID)
}

func TestBudgetRepository_Update(t *testing.T) {
	repo, userID, categoryID := newBudgetRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Budget{
		UserID: userID, CategoryID: categoryID, LimitCents: 10000,
		Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD", NotifyAtPercent: 50,
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, &domain.Budget{
		ID:              created.ID,
		UserID:          userID,
		CategoryID:      categoryID,
		LimitCents:      30000,
		Period:          domain.BudgetPeriodWeekly,
		NotifyAtPercent: 90,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(30000), updated.LimitCents)
	assert.Equal(t, 90, updated.NotifyAtPercent)
}

func TestBudgetRepository_Delete(t *testing.T) {
	repo, userID, categoryID := newBudgetRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Budget{
		UserID: userID, CategoryID: categoryID, LimitCents: 1000,
		Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, created.ID, userID))
	_, err = repo.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, domain.ErrBudgetNotFound)
}

func TestBudgetRepository_UpdateLastNotified(t *testing.T) {
	repo, userID, categoryID := newBudgetRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Budget{
		UserID: userID, CategoryID: categoryID, LimitCents: 1000,
		Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD", NotifyAtPercent: 50,
	})
	require.NoError(t, err)

	require.NoError(t, repo.UpdateLastNotified(ctx, created.ID, 75))

	fetched, err := repo.GetByID(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, 75, fetched.LastNotifiedPercent)
}

func TestBudgetRepository_ListDistinctUserIDs(t *testing.T) {
	repo, user1, cat1 := newBudgetRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	user2 := testutil.SeedUser(t, pool, time.Now().UnixNano()+2)
	cat2 := testutil.SeedCategory(t, pool, user2, "X", "expense")

	ctx := context.Background()
	_, _ = repo.Create(ctx, &domain.Budget{UserID: user1, CategoryID: cat1, LimitCents: 1, Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD"})
	_, _ = repo.Create(ctx, &domain.Budget{UserID: user1, CategoryID: cat1, LimitCents: 2, Period: domain.BudgetPeriodWeekly, CurrencyCode: "USD"})
	_, _ = repo.Create(ctx, &domain.Budget{UserID: user2, CategoryID: cat2, LimitCents: 3, Period: domain.BudgetPeriodMonthly, CurrencyCode: "USD"})

	ids, err := repo.ListDistinctUserIDs(ctx)
	require.NoError(t, err)
	assert.Contains(t, ids, user1)
	assert.Contains(t, ids, user2)
}
