package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func newSnapshotService(repo *mocks.MockSnapshotStorer, provider *mocks.MockRateProvider) *service.SnapshotService {
	return service.NewSnapshotService(repo, provider, testutil.TestLogger())
}

func TestSnapshotService_GetRate_SameCurrencyReturnsOne(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	svc := newSnapshotService(repo, &mocks.MockRateProvider{})

	rate, err := svc.GetRate(context.Background(), time.Now(), "USD", "USD")
	require.NoError(t, err)
	assert.Equal(t, 1.0, rate)
}

func TestSnapshotService_GetRate_LooksUpHistoricalRate(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	svc := newSnapshotService(repo, &mocks.MockRateProvider{})
	date := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	repo.On("GetRateOrLatest", mock.Anything, date, "USD", "EUR").Return(0.92, nil)

	rate, err := svc.GetRate(context.Background(), date, "USD", "EUR")
	require.NoError(t, err)
	assert.InDelta(t, 0.92, rate, 0.0001)
}

func TestSnapshotService_GetRate_WrapsRepoErrorAsExchangeUnavailable(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	svc := newSnapshotService(repo, &mocks.MockRateProvider{})
	repo.On("GetRateOrLatest", mock.Anything, mock.Anything, "USD", "EUR").
		Return(0.0, errors.New("not in db"))

	_, err := svc.GetRate(context.Background(), time.Now(), "USD", "EUR")
	assert.ErrorIs(t, err, domain.ErrExchangeRateUnavailable)
}

func TestSnapshotService_Convert_AppliesRateAndRounds(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	svc := newSnapshotService(repo, &mocks.MockRateProvider{})
	repo.On("GetRateOrLatest", mock.Anything, mock.Anything, "USD", "EUR").Return(0.92, nil)

	got, err := svc.Convert(context.Background(), 10000, "USD", "EUR", time.Now())
	require.NoError(t, err)
	assert.Equal(t, int64(9200), got)
}

func TestSnapshotService_SaveDaily_NoBasesNoFetch(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	provider := &mocks.MockRateProvider{}
	svc := newSnapshotService(repo, provider)

	repo.On("ListDistinctBaseCurrencies", mock.Anything).Return([]string{}, nil)

	err := svc.SaveDaily(context.Background())
	require.NoError(t, err)
	provider.AssertNotCalled(t, "FetchRates", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSnapshotService_SaveDaily_FetchesAndUpserts(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	provider := &mocks.MockRateProvider{}
	svc := newSnapshotService(repo, provider)

	repo.On("ListDistinctBaseCurrencies", mock.Anything).Return([]string{"USD"}, nil)
	provider.On("FetchRates", mock.Anything, "USD").Return(map[string]float64{
		"USD": 1.0,
		"EUR": 0.92,
		"GBP": 0.79,
	}, nil)
	// Upsert called for non-base entries (EUR, GBP)
	repo.On("Upsert", mock.Anything, mock.AnythingOfType("time.Time"), "USD", "EUR", 0.92).Return(nil)
	repo.On("Upsert", mock.Anything, mock.AnythingOfType("time.Time"), "USD", "GBP", 0.79).Return(nil)

	err := svc.SaveDaily(context.Background())
	require.NoError(t, err)
	repo.AssertNumberOfCalls(t, "Upsert", 2)
}

func TestSnapshotService_SaveDaily_ContinuesOnFetchError(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	provider := &mocks.MockRateProvider{}
	svc := newSnapshotService(repo, provider)

	repo.On("ListDistinctBaseCurrencies", mock.Anything).Return([]string{"USD", "EUR"}, nil)
	provider.On("FetchRates", mock.Anything, "USD").Return(nil, errors.New("API down"))
	provider.On("FetchRates", mock.Anything, "EUR").Return(map[string]float64{"USD": 1.08}, nil)
	repo.On("Upsert", mock.Anything, mock.AnythingOfType("time.Time"), "EUR", "USD", 1.08).Return(nil)

	err := svc.SaveDaily(context.Background())
	require.NoError(t, err)
	repo.AssertNumberOfCalls(t, "Upsert", 1)
}

func TestSnapshotService_SaveDaily_ListBasesError(t *testing.T) {
	repo := &mocks.MockSnapshotStorer{}
	svc := newSnapshotService(repo, &mocks.MockRateProvider{})
	repo.On("ListDistinctBaseCurrencies", mock.Anything).Return(nil, errors.New("db down"))

	err := svc.SaveDaily(context.Background())
	require.Error(t, err)
}
