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

func newSavingsRepoFixtures(t *testing.T) (*repository.SavingsGoalRepository, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	return repository.NewSavingsGoalRepository(pool), userID
}

func TestSavingsGoalRepository_Create_PersistsAttributes(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()
	deadline := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second)

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID:       userID,
		Name:         "Vacation",
		TargetCents:  500_000,
		CurrencyCode: "USD",
		Deadline:     &deadline,
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, "Vacation", created.Name)
	assert.Equal(t, int64(500_000), created.TargetCents)
	assert.Equal(t, int64(0), created.CurrentCents)
	require.NotNil(t, created.Deadline)
}

func TestSavingsGoalRepository_GetByID_NotFound(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99_999, userID)
	assert.ErrorIs(t, err, domain.ErrGoalNotFound)
}

func TestSavingsGoalRepository_Deposit_IncreasesCurrent(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID: userID, Name: "Car", TargetCents: 1_000_000, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	updated, err := repo.Deposit(ctx, created.ID, userID, 25_000)
	require.NoError(t, err)
	assert.Equal(t, int64(25_000), updated.CurrentCents)

	updated, err = repo.Deposit(ctx, created.ID, userID, 75_000)
	require.NoError(t, err)
	assert.Equal(t, int64(100_000), updated.CurrentCents)
}

func TestSavingsGoalRepository_Withdraw_DecreasesCurrent(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID: userID, Name: "Phone", TargetCents: 100_000, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	_, err = repo.Deposit(ctx, created.ID, userID, 50_000)
	require.NoError(t, err)

	updated, err := repo.Withdraw(ctx, created.ID, userID, 20_000)
	require.NoError(t, err)
	assert.Equal(t, int64(30_000), updated.CurrentCents)
}

func TestSavingsGoalRepository_Withdraw_RejectsInsufficient(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID: userID, Name: "Phone", TargetCents: 100_000, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	_, err = repo.Withdraw(ctx, created.ID, userID, 1)
	assert.ErrorIs(t, err, domain.ErrInsufficientGoalFunds)
}

func TestSavingsGoalRepository_ListHistory(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID: userID, Name: "Vacation", TargetCents: 100_000, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	_, err = repo.Deposit(ctx, created.ID, userID, 1000)
	require.NoError(t, err)
	_, err = repo.Deposit(ctx, created.ID, userID, 500)
	require.NoError(t, err)
	_, err = repo.Withdraw(ctx, created.ID, userID, 200)
	require.NoError(t, err)

	history, err := repo.ListHistory(ctx, created.ID, userID)
	require.NoError(t, err)
	require.Len(t, history, 3)
}

func TestSavingsGoalRepository_Update_AndDelete(t *testing.T) {
	repo, userID := newSavingsRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.SavingsGoal{
		UserID: userID, Name: "Old", TargetCents: 100, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, &domain.SavingsGoal{
		ID:           created.ID,
		UserID:       userID,
		Name:         "New",
		TargetCents:  200,
		CurrencyCode: "USD",
	})
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, int64(200), updated.TargetCents)

	require.NoError(t, repo.Delete(ctx, created.ID, userID))
	_, err = repo.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, domain.ErrGoalNotFound)
}
