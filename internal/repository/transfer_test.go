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

func newTransferRepoFixtures(t *testing.T) (*repository.TransferRepository, int64, int64, int64) {
	t.Helper()
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	userID := testutil.SeedUser(t, pool, time.Now().UnixNano())
	from := testutil.SeedAccount(t, pool, userID, "USD")
	to := testutil.SeedAccount(t, pool, userID, "EUR")
	return repository.NewTransferRepository(pool), userID, from, to
}

func TestTransferRepository_Create_PersistsAttributes(t *testing.T) {
	repo, userID, from, to := newTransferRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Transfer{
		UserID:           userID,
		FromAccountID:    from,
		ToAccountID:      to,
		AmountCents:      10_000,
		FromCurrencyCode: "USD",
		ToCurrencyCode:   "EUR",
		ExchangeRate:     0.92,
		Note:             "March transfer",
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, int64(10_000), created.AmountCents)
	assert.Equal(t, "USD", created.FromCurrencyCode)
	assert.Equal(t, "EUR", created.ToCurrencyCode)
	assert.InDelta(t, 0.92, created.ExchangeRate, 0.0001)
}

func TestTransferRepository_GetByID_NotFound(t *testing.T) {
	repo, userID, _, _ := newTransferRepoFixtures(t)
	_, err := repo.GetByID(context.Background(), 99_999, userID)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}

func TestTransferRepository_ListByUser_Pagination(t *testing.T) {
	repo, userID, from, to := newTransferRepoFixtures(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := repo.Create(ctx, &domain.Transfer{
			UserID: userID, FromAccountID: from, ToAccountID: to,
			AmountCents: int64(100 * (i + 1)), FromCurrencyCode: "USD", ToCurrencyCode: "EUR",
			ExchangeRate: 0.9,
		})
		require.NoError(t, err)
	}

	page1, err := repo.ListByUser(ctx, userID, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page1, 2)

	page2, err := repo.ListByUser(ctx, userID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, page2, 1)
}

func TestTransferRepository_Count(t *testing.T) {
	repo, userID, from, to := newTransferRepoFixtures(t)
	ctx := context.Background()

	for i := 0; i < 4; i++ {
		_, err := repo.Create(ctx, &domain.Transfer{
			UserID: userID, FromAccountID: from, ToAccountID: to,
			AmountCents: 100, FromCurrencyCode: "USD", ToCurrencyCode: "EUR", ExchangeRate: 1,
		})
		require.NoError(t, err)
	}

	n, err := repo.Count(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, int64(4), n)
}

func TestTransferRepository_ListByAccount(t *testing.T) {
	repo, userID, from, to := newTransferRepoFixtures(t)
	pool := testutil.OpenTestPool(t)
	other := testutil.SeedAccount(t, pool, userID, "USD")
	ctx := context.Background()

	_, err := repo.Create(ctx, &domain.Transfer{
		UserID: userID, FromAccountID: from, ToAccountID: to,
		AmountCents: 100, FromCurrencyCode: "USD", ToCurrencyCode: "EUR", ExchangeRate: 0.9,
	})
	require.NoError(t, err)
	_, err = repo.Create(ctx, &domain.Transfer{
		UserID: userID, FromAccountID: from, ToAccountID: other,
		AmountCents: 200, FromCurrencyCode: "USD", ToCurrencyCode: "USD", ExchangeRate: 1,
	})
	require.NoError(t, err)

	list, err := repo.ListByAccount(ctx, userID, from)
	require.NoError(t, err)
	require.Len(t, list, 2)
}

func TestTransferRepository_Delete(t *testing.T) {
	repo, userID, from, to := newTransferRepoFixtures(t)
	ctx := context.Background()

	created, err := repo.Create(ctx, &domain.Transfer{
		UserID: userID, FromAccountID: from, ToAccountID: to,
		AmountCents: 100, FromCurrencyCode: "USD", ToCurrencyCode: "EUR", ExchangeRate: 0.9,
	})
	require.NoError(t, err)

	require.NoError(t, repo.Delete(ctx, created.ID, userID))
	_, err = repo.GetByID(ctx, created.ID, userID)
	assert.ErrorIs(t, err, domain.ErrTransactionNotFound)
}
