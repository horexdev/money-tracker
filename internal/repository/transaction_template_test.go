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

func newTemplateRepoFixtures(t *testing.T) (*repository.TransactionTemplateRepository, int64, int64, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	accountID := testutil.SeedAccount(t, pool, userID, "USD")
	categoryID := testutil.SeedCategory(t, pool, userID, "Food", "expense")
	return repository.NewTransactionTemplateRepository(pool), userID, accountID, categoryID
}

func TestTemplateRepository_Create_PersistsAllFields(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.TransactionTemplate{
		UserID:       userID,
		Name:         "Coffee",
		Type:         domain.TransactionTypeExpense,
		AmountCents:  30000,
		AmountFixed:  true,
		CategoryID:   categoryID,
		AccountID:    accountID,
		CurrencyCode: "USD",
		Note:         "morning",
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, "Coffee", created.Name)
	assert.Equal(t, int64(30000), created.AmountCents)
	assert.True(t, created.AmountFixed)
	assert.Equal(t, int32(0), created.SortOrder, "first template gets sort_order=0")
	// Joined fields populated.
	assert.Equal(t, "Food", created.CategoryName)
	assert.Equal(t, "package", created.CategoryIcon)
	assert.Equal(t, "#6366f1", created.CategoryColor)
}

func TestTemplateRepository_Create_AutoIncrementsSortOrder(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &domain.TransactionTemplate{
			UserID: userID, Name: "T", Type: domain.TransactionTypeExpense,
			AmountCents: 100, AmountFixed: true, CategoryID: categoryID,
			AccountID: accountID, CurrencyCode: "USD",
		})
		require.NoError(t, err)
	}

	list, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, int32(0), list[0].SortOrder)
	assert.Equal(t, int32(1), list[1].SortOrder)
	assert.Equal(t, int32(2), list[2].SortOrder)
}

func TestTemplateRepository_GetByID_NotFound(t *testing.T) {
	repo, userID, _, _ := newTemplateRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99999, userID)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateRepository_GetByID_ScopedToUser(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.TransactionTemplate{
		UserID: userID, Name: "X", Type: domain.TransactionTypeExpense,
		AmountCents: 100, AmountFixed: true, CategoryID: categoryID,
		AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	otherUserID := userID + 1
	_, err = repo.GetByID(ctx, created.ID, otherUserID)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateRepository_Update_PersistsChanges(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.TransactionTemplate{
		UserID: userID, Name: "Old", Type: domain.TransactionTypeExpense,
		AmountCents: 100, AmountFixed: true, CategoryID: categoryID,
		AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	created.Name = "New"
	created.AmountCents = 5000
	created.AmountFixed = false
	updated, err := repo.Update(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "New", updated.Name)
	assert.Equal(t, int64(5000), updated.AmountCents)
	assert.False(t, updated.AmountFixed)
}

func TestTemplateRepository_Update_NotFound(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	_, err := repo.Update(context.Background(), &domain.TransactionTemplate{
		ID: 99999, UserID: userID, Type: domain.TransactionTypeExpense,
		AmountCents: 1, CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
	})
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateRepository_Delete_RemovesScopedToUser(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.TransactionTemplate{
		UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 1,
		AmountFixed: true, CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, created.ID, userID))

	_, err = repo.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateRepository_Delete_NotFound(t *testing.T) {
	repo, userID, _, _ := newTemplateRepoFixtures(t)
	err := repo.Delete(context.Background(), 99999, userID)
	assert.ErrorIs(t, err, domain.ErrTemplateNotFound)
}

func TestTemplateRepository_Reorder_AppliesNewOrder(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	ids := make([]int64, 0, 3)
	for _, name := range []string{"A", "B", "C"} {
		c, err := repo.Create(ctx, &domain.TransactionTemplate{
			UserID: userID, Name: name, Type: domain.TransactionTypeExpense,
			AmountCents: 1, AmountFixed: true, CategoryID: categoryID,
			AccountID: accountID, CurrencyCode: "USD",
		})
		require.NoError(t, err)
		ids = append(ids, c.ID)
	}

	// Reverse the order: C, B, A.
	require.NoError(t, repo.Reorder(ctx, userID, []int64{ids[2], ids[1], ids[0]}))

	list, err := repo.ListByUser(ctx, userID)
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "C", list[0].Name)
	assert.Equal(t, "B", list[1].Name)
	assert.Equal(t, "A", list[2].Name)
}

func TestTemplateRepository_Reorder_IgnoresOtherUserIDs(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.TransactionTemplate{
		UserID: userID, Name: "Mine", Type: domain.TransactionTypeExpense,
		AmountCents: 1, AmountFixed: true, CategoryID: categoryID,
		AccountID: accountID, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	// Foreign id 99999 is silently skipped; own template still gets sort_order=1.
	require.NoError(t, repo.Reorder(ctx, userID, []int64{99999, created.ID}))

	got, err := repo.GetByID(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, int32(1), got.SortOrder)
}

func TestTemplateRepository_CountByAccount(t *testing.T) {
	repo, userID, accountID, categoryID := newTemplateRepoFixtures(t)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		_, err := repo.Create(ctx, &domain.TransactionTemplate{
			UserID: userID, Type: domain.TransactionTypeExpense, AmountCents: 1,
			AmountFixed: true, CategoryID: categoryID, AccountID: accountID, CurrencyCode: "USD",
		})
		require.NoError(t, err)
	}

	n, err := repo.CountByAccount(ctx, accountID)
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
}
