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

func newAccountRepoFixtures(t *testing.T) (*repository.AccountRepository, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	return repository.NewAccountRepository(pool), userID
}

func TestAccountRepository_CreateAndGetByID(t *testing.T) {
	repo, userID := newAccountRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Account{
		UserID:         userID,
		Name:           "Main",
		Icon:           "wallet",
		Color:          "#6366f1",
		Type:           domain.AccountTypeChecking,
		CurrencyCode:   "USD",
		IsDefault:      true,
		IncludeInTotal: true,
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)

	fetched, err := repo.GetByID(ctx, created.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "Main", fetched.Name)
	assert.Equal(t, "USD", fetched.CurrencyCode)
	assert.True(t, fetched.IsDefault)
	assert.Equal(t, domain.AccountTypeChecking, fetched.Type)
}

func TestAccountRepository_GetByID_OtherUser_ReturnsErrAccountNotFound(t *testing.T) {
	repo, owner := newAccountRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	stranger := testutil.SeedUser(t, pool, time.Now().UnixNano()+1)

	ctx := context.Background()
	created, err := repo.Create(ctx, &domain.Account{
		UserID: owner, Name: "Main", Icon: "wallet", Color: "#000",
		Type: domain.AccountTypeChecking, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, created.ID, stranger)
	assert.ErrorIs(t, err, domain.ErrAccountNotFound)
}

func TestAccountRepository_ListByUser_OnlyOwnAccounts(t *testing.T) {
	repo, user1 := newAccountRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	user2 := testutil.SeedUser(t, pool, time.Now().UnixNano()+2)

	ctx := context.Background()
	for i, name := range []string{"Cash", "Card"} {
		_, err := repo.Create(ctx, &domain.Account{
			UserID: user1, Name: name, Icon: "wallet", Color: "#000",
			Type: domain.AccountTypeChecking, CurrencyCode: "USD",
			IsDefault: i == 0,
		})
		require.NoError(t, err)
	}
	_, err := repo.Create(ctx, &domain.Account{
		UserID: user2, Name: "Other", Icon: "wallet", Color: "#000",
		Type: domain.AccountTypeChecking, CurrencyCode: "EUR",
	})
	require.NoError(t, err)

	list, err := repo.ListByUser(ctx, user1)
	require.NoError(t, err)
	require.Len(t, list, 2)
	for _, a := range list {
		assert.Equal(t, user1, a.UserID)
	}
}

func TestAccountRepository_SetDefault_MovesFlag(t *testing.T) {
	repo, userID := newAccountRepoFixtures(t)
	ctx := context.Background()

	a1, err := repo.Create(ctx, &domain.Account{
		UserID: userID, Name: "First", Icon: "wallet", Color: "#000",
		Type: domain.AccountTypeChecking, CurrencyCode: "USD", IsDefault: true,
	})
	require.NoError(t, err)
	a2, err := repo.Create(ctx, &domain.Account{
		UserID: userID, Name: "Second", Icon: "wallet", Color: "#000",
		Type: domain.AccountTypeChecking, CurrencyCode: "USD",
	})
	require.NoError(t, err)

	switched, err := repo.SetDefault(ctx, a2.ID, userID)
	require.NoError(t, err)
	assert.True(t, switched.IsDefault)

	first, err := repo.GetByID(ctx, a1.ID, userID)
	require.NoError(t, err)
	assert.False(t, first.IsDefault, "old default must be cleared")
}

func TestAccountRepository_Update_KeepsCurrencyImmutable(t *testing.T) {
	repo, userID := newAccountRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Account{
		UserID: userID, Name: "Main", Icon: "wallet", Color: "#111",
		Type: domain.AccountTypeChecking, CurrencyCode: "USD", IncludeInTotal: true,
	})
	require.NoError(t, err)

	updated, err := repo.Update(ctx, &domain.Account{
		ID: created.ID, UserID: userID,
		Name: "Renamed", Icon: "card", Color: "#222",
		Type:           domain.AccountTypeCash,
		CurrencyCode:   "EUR", // ignored — currency is immutable
		IncludeInTotal: false,
	})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Name)
	assert.Equal(t, "card", updated.Icon)
	assert.Equal(t, "USD", updated.CurrencyCode, "currency_code must stay USD")
	assert.False(t, updated.IncludeInTotal)
}
