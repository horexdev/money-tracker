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

func newCategoryRepoFixtures(t *testing.T) (*repository.CategoryRepository, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	return repository.NewCategoryRepository(pool), userID
}

func TestCategoryRepository_CreateForUser_PersistsAttributes(t *testing.T) {
	repo, userID := newCategoryRepoFixtures(t)
	ctx := context.Background()

	cat, err := repo.CreateForUser(ctx, userID, "Coffee", "coffee", "expense", "#a87654")
	require.NoError(t, err)
	require.NotZero(t, cat.ID)
	assert.Equal(t, "Coffee", cat.Name)
	assert.Equal(t, "coffee", cat.Icon)
	assert.Equal(t, domain.CategoryType("expense"), cat.Type)
	assert.Equal(t, "#a87654", cat.Color)
}

func TestCategoryRepository_ListForUser_ScopedToUser(t *testing.T) {
	repo, user1 := newCategoryRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	user2 := testutil.SeedUser(t, pool, time.Now().UnixNano()+1)

	ctx := context.Background()
	_, err := repo.CreateForUser(ctx, user1, "Food", "food", "expense", "#fff")
	require.NoError(t, err)
	_, err = repo.CreateForUser(ctx, user2, "Travel", "globe", "expense", "#000")
	require.NoError(t, err)

	cats, err := repo.ListForUser(ctx, user1)
	require.NoError(t, err)
	for _, c := range cats {
		assert.NotEqual(t, "Travel", c.Name, "user1 must not see user2 categories")
	}
	hasFood := false
	for _, c := range cats {
		if c.Name == "Food" {
			hasFood = true
		}
	}
	assert.True(t, hasFood, "user1 must see their own Food category")
}

func TestCategoryRepository_GetByName_ReturnsErrCategoryNotFound(t *testing.T) {
	repo, userID := newCategoryRepoFixtures(t)
	ctx := context.Background()

	_, err := repo.GetByName(ctx, userID, "NonExistent")
	assert.ErrorIs(t, err, domain.ErrCategoryNotFound)
}

func TestCategoryRepository_Update_ChangesMutableFields(t *testing.T) {
	repo, userID := newCategoryRepoFixtures(t)
	ctx := context.Background()

	cat, err := repo.CreateForUser(ctx, userID, "Food", "food", "expense", "#fff")
	require.NoError(t, err)

	updated, err := repo.Update(ctx, userID, cat.ID, "Groceries", "shopping-cart", "expense", "#abc")
	require.NoError(t, err)
	assert.Equal(t, "Groceries", updated.Name)
	assert.Equal(t, "shopping-cart", updated.Icon)
	assert.Equal(t, "#abc", updated.Color)
}

func TestCategoryRepository_SoftDelete_HidesFromList(t *testing.T) {
	repo, userID := newCategoryRepoFixtures(t)
	ctx := context.Background()

	cat, err := repo.CreateForUser(ctx, userID, "Hobbies", "music-note", "expense", "#fff")
	require.NoError(t, err)

	require.NoError(t, repo.SoftDelete(ctx, cat.ID, userID))

	cats, err := repo.ListForUser(ctx, userID)
	require.NoError(t, err)
	for _, c := range cats {
		assert.NotEqual(t, cat.ID, c.ID, "soft-deleted category must not appear in ListForUser")
	}
}
