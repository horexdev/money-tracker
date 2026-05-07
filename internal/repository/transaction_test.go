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

func newTransactionFixtures(t *testing.T) (*repository.TransactionRepository, int64, int64, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	accountID := testutil.SeedAccount(t, pool, userID, "USD")
	categoryID := testutil.SeedCategory(t, pool, userID, "Food", "expense")

	return repository.NewTransactionRepository(pool), userID, accountID, categoryID
}

func TestTransactionRepository_Create_PersistsAllFields(t *testing.T) {
	repo, userID, accountID, categoryID := newTransactionFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Transaction{
		UserID:       userID,
		Type:         domain.TransactionTypeExpense,
		AmountCents:  1500,
		CategoryID:   categoryID,
		AccountID:    accountID,
		Note:         "Lunch",
		CurrencyCode: "USD",
		SnapshotDate: time.Now().UTC().Truncate(24 * time.Hour),
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, userID, created.UserID)
	assert.Equal(t, domain.TransactionTypeExpense, created.Type)
	assert.Equal(t, int64(1500), created.AmountCents)
	assert.Equal(t, categoryID, created.CategoryID)
	assert.Equal(t, accountID, created.AccountID)
	assert.Equal(t, "USD", created.CurrencyCode)
	assert.Equal(t, "Lunch", created.Note)
	assert.False(t, created.IsAdjustment)
}

func TestTransactionRepository_List_ScopedToUser(t *testing.T) {
	repo, user1, acct1, cat1 := newTransactionFixtures(t)
	pool := testutil.OpenTestPool(t)
	user2 := testutil.SeedUser(t, pool, time.Now().UnixNano()+1)
	acct2 := testutil.SeedAccount(t, pool, user2, "USD")
	cat2 := testutil.SeedCategory(t, pool, user2, "Food", "expense")

	ctx := context.Background()
	_, err := repo.Create(ctx, &domain.Transaction{
		UserID: user1, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: cat1, AccountID: acct1, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: user2, Type: domain.TransactionTypeExpense, AmountCents: 200,
		CategoryID: cat2, AccountID: acct2, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	list1, err := repo.List(ctx, user1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list1, 1)
	assert.Equal(t, int64(100), list1[0].AmountCents)

	list2, err := repo.List(ctx, user2, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list2, 1)
	assert.Equal(t, int64(200), list2[0].AmountCents)
}

func TestTransactionRepository_Delete_ScopedToUser(t *testing.T) {
	repo, owner, acct, cat := newTransactionFixtures(t)
	pool := testutil.OpenTestPool(t)
	other := testutil.SeedUser(t, pool, time.Now().UnixNano()+2)

	ctx := context.Background()
	created, err := repo.Create(ctx, &domain.Transaction{
		UserID: owner, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: cat, AccountID: acct, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	// Delete attempted by a different user — must be a no-op.
	require.NoError(t, repo.Delete(ctx, created.ID, other))

	count, err := repo.Count(ctx, owner)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "transaction must still belong to the owner")

	// Owner can delete their own row.
	require.NoError(t, repo.Delete(ctx, created.ID, owner))
	count, err = repo.Count(ctx, owner)
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestTransactionRepository_GetBalance_SumsByType(t *testing.T) {
	repo, userID, accountID, categoryID := newTransactionFixtures(t)
	pool := testutil.OpenTestPool(t)
	incomeCat := testutil.SeedCategory(t, pool, userID, "Salary", "income")

	ctx := context.Background()
	_, err := repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 300,
		CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 200,
		CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeIncome, AmountCents: 1000,
		CategoryID: incomeCat, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	income, expense, err := repo.GetBalance(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), income)
	assert.Equal(t, int64(500), expense)
}

func TestTransactionRepository_ListByCategoryWithDateRange_FiltersByCategory(t *testing.T) {
	repo, userID, accountID, foodCat := newTransactionFixtures(t)
	pool := testutil.OpenTestPool(t)
	otherCat := testutil.SeedCategory(t, pool, userID, "Travel", "expense")

	ctx := context.Background()
	_, err := repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: foodCat, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 200,
		CategoryID: foodCat, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 999,
		CategoryID: otherCat, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	list, err := repo.ListByCategoryWithDateRange(ctx, userID, foodCat, nil, nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	for _, tx := range list {
		assert.Equal(t, foodCat, tx.CategoryID)
	}

	count, err := repo.CountByCategoryWithDateRange(ctx, userID, foodCat, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestTransactionRepository_ListByCategoryWithDateRange_RespectsDateRange(t *testing.T) {
	repo, userID, accountID, categoryID := newTransactionFixtures(t)
	ctx := context.Background()

	older := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)

	_, err := repo.CreateWithDate(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
		CreatedAt: older, SnapshotDate: older.Truncate(24 * time.Hour),
	})
	require.NoError(t, err)
	_, err = repo.CreateWithDate(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 200,
		CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
		CreatedAt: newer, SnapshotDate: newer.Truncate(24 * time.Hour),
	})
	require.NoError(t, err)

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 30, 23, 59, 59, 0, time.UTC)
	list, err := repo.ListByCategoryWithDateRange(ctx, userID, categoryID, &from, &to, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, int64(200), list[0].AmountCents)
}

func TestTransactionRepository_ListByAccountAndCategoryWithDateRange_FiltersByBoth(t *testing.T) {
	repo, userID, account1, categoryID := newTransactionFixtures(t)
	pool := testutil.OpenTestPool(t)
	account2 := testutil.SeedAccount(t, pool, userID, "USD")

	ctx := context.Background()
	_, err := repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: categoryID, AccountID: account1, CurrencyCode: "USD",
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 200,
		CategoryID: categoryID, AccountID: account2, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	list, err := repo.ListByAccountAndCategoryWithDateRange(ctx, userID, account1, categoryID, nil, nil, 10, 0)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, account1, list[0].AccountID)
	assert.Equal(t, int64(100), list[0].AmountCents)
}

func TestTransactionRepository_StatsByCategory_IncludesCategoryID(t *testing.T) {
	repo, userID, accountID, categoryID := newTransactionFixtures(t)
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 1500,
		CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	stats, err := repo.StatsByCategory(ctx, userID, time.Now().Add(-24*time.Hour), time.Now().Add(24*time.Hour))
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, categoryID, stats[0].CategoryID)
	assert.Equal(t, "Food", stats[0].CategoryName)
}

func TestTransactionRepository_Update_ChangesMutableFields(t *testing.T) {
	repo, userID, accountID, categoryID := newTransactionFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Transaction{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 100,
		CategoryID: categoryID, AccountID: accountID, Note: "original", CurrencyCode: "USD",
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, &domain.Transaction{
		ID:          created.ID,
		UserID:      userID,
		AmountCents: 250,
		CategoryID:  categoryID,
		Note:        "amended",
		CreatedAt:   created.CreatedAt,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(250), updated.AmountCents)
	assert.Equal(t, "amended", updated.Note)
}
