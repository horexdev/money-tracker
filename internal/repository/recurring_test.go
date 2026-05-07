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

func newRecurringRepoFixtures(t *testing.T) (*repository.RecurringRepository, int64, int64, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	accountID := testutil.SeedAccount(t, pool, userID, "USD")
	categoryID := testutil.SeedCategory(t, pool, userID, "Subscriptions", "expense")
	return repository.NewRecurringRepository(pool), userID, accountID, categoryID
}

func TestRecurringRepository_Create_PersistsAttributes(t *testing.T) {
	repo, userID, accountID, categoryID := newRecurringRepoFixtures(t)
	ctx := context.Background()
	nextRun := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	created, err := repo.Create(ctx, &domain.RecurringTransaction{
		UserID:       userID,
		AccountID:    accountID,
		CategoryID:   categoryID,
		Type:         domain.TransactionTypeExpense,
		AmountCents:  9999,
		CurrencyCode: "USD",
		Note:         "Netflix",
		Frequency:    domain.FrequencyMonthly,
		NextRunAt:    nextRun,
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, int64(9999), created.AmountCents)
	assert.Equal(t, domain.FrequencyMonthly, created.Frequency)
	assert.True(t, created.IsActive)
}

func TestRecurringRepository_GetByID_NotFound(t *testing.T) {
	repo, userID, _, _ := newRecurringRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99_999, userID)
	assert.ErrorIs(t, err, domain.ErrRecurringNotFound)
}

func TestRecurringRepository_ToggleActive(t *testing.T) {
	repo, userID, accountID, categoryID := newRecurringRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.RecurringTransaction{
		UserID: userID, AccountID: accountID, CategoryID: categoryID,
		Type: domain.TransactionTypeExpense, AmountCents: 1, CurrencyCode: "USD",
		Frequency: domain.FrequencyMonthly, NextRunAt: time.Now().UTC().Add(time.Hour),
	})
	require.NoError(t, err)
	assert.True(t, created.IsActive)

	toggled, err := repo.ToggleActive(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.False(t, toggled.IsActive)

	toggledBack, err := repo.ToggleActive(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.True(t, toggledBack.IsActive)
}

func TestRecurringRepository_GetDue_ReturnsOnlyOverdueAndActive(t *testing.T) {
	repo, userID, accountID, categoryID := newRecurringRepoFixtures(t)
	ctx := context.Background()

	past := time.Now().UTC().Add(-time.Hour)
	future := time.Now().UTC().Add(24 * time.Hour)

	overdue, err := repo.Create(ctx, &domain.RecurringTransaction{
		UserID: userID, AccountID: accountID, CategoryID: categoryID,
		Type: domain.TransactionTypeExpense, AmountCents: 1, CurrencyCode: "USD",
		Frequency: domain.FrequencyDaily, NextRunAt: past, Note: "due",
	})
	require.NoError(t, err)

	_, err = repo.Create(ctx, &domain.RecurringTransaction{
		UserID: userID, AccountID: accountID, CategoryID: categoryID,
		Type: domain.TransactionTypeExpense, AmountCents: 2, CurrencyCode: "USD",
		Frequency: domain.FrequencyDaily, NextRunAt: future, Note: "later",
	})
	require.NoError(t, err)

	due, err := repo.GetDue(ctx, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, due, 1)
	assert.Equal(t, overdue.ID, due[0].ID)
}

func TestRecurringRepository_UpdateNextRun(t *testing.T) {
	repo, userID, accountID, categoryID := newRecurringRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.RecurringTransaction{
		UserID: userID, AccountID: accountID, CategoryID: categoryID,
		Type: domain.TransactionTypeExpense, AmountCents: 1, CurrencyCode: "USD",
		Frequency: domain.FrequencyMonthly, NextRunAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	newRun := time.Now().UTC().Add(7 * 24 * time.Hour).Truncate(time.Second)
	require.NoError(t, repo.UpdateNextRun(ctx, created.ID, newRun))

	fetched, err := repo.GetByID(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.WithinDuration(t, newRun, fetched.NextRunAt, time.Second)
}

func TestRecurringRepository_Delete(t *testing.T) {
	repo, userID, accountID, categoryID := newRecurringRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.RecurringTransaction{
		UserID: userID, AccountID: accountID, CategoryID: categoryID,
		Type: domain.TransactionTypeExpense, AmountCents: 1, CurrencyCode: "USD",
		Frequency: domain.FrequencyMonthly, NextRunAt: time.Now().UTC(),
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, created.ID, userID))
	_, err = repo.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, domain.ErrRecurringNotFound)
}
