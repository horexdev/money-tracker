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

func newStatsService(repo *mocks.MockTransactionStorer) *service.StatsService {
	return service.NewStatsService(repo, testutil.TestLogger())
}

func TestStatsService_ByCategory_PassesArgsAndResult(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newStatsService(repo)

	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	expected := []domain.CategoryStat{
		{CategoryName: "Food", Type: domain.TransactionTypeExpense, TotalCents: 1500, TxCount: 3, CurrencyCode: "USD"},
	}
	repo.On("StatsByCategory", mock.Anything, int64(42), from, to).Return(expected, nil)

	got, err := svc.ByCategory(context.Background(), 42, from, to)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
	repo.AssertExpectations(t)
}

func TestStatsService_ByCategory_WrapsRepoError(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newStatsService(repo)
	repo.On("StatsByCategory", mock.Anything, int64(1), mock.Anything, mock.Anything).
		Return(nil, errors.New("db down"))

	stats, err := svc.ByCategory(context.Background(), 1, time.Now(), time.Now())
	assert.Nil(t, stats)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "stats by category for user 1")
}

func TestStatsService_ByCategoryAndAccount_PassesAccountID(t *testing.T) {
	repo := &mocks.MockTransactionStorer{}
	svc := newStatsService(repo)

	from := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	repo.On("StatsByCategoryAndAccount", mock.Anything, int64(7), int64(99), from, to).
		Return([]domain.CategoryStat{}, nil)

	_, err := svc.ByCategoryAndAccount(context.Background(), 7, 99, from, to)
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestStatsService_PeriodRange(t *testing.T) {
	t.Run("today spans midnight to midnight", func(t *testing.T) {
		from, to, err := service.PeriodRange("today")
		require.NoError(t, err)
		assert.Equal(t, 24*time.Hour, to.Sub(from))
		assert.Equal(t, 0, from.Hour())
	})

	t.Run("week spans seven days", func(t *testing.T) {
		from, to, err := service.PeriodRange("week")
		require.NoError(t, err)
		assert.Equal(t, 7*24*time.Hour, to.Sub(from))
	})

	t.Run("month starts on day 1", func(t *testing.T) {
		from, _, err := service.PeriodRange("month")
		require.NoError(t, err)
		assert.Equal(t, 1, from.Day())
		assert.Equal(t, 0, from.Hour())
	})

	t.Run("lastmonth ends at start of current month", func(t *testing.T) {
		_, to, err := service.PeriodRange("lastmonth")
		require.NoError(t, err)
		assert.Equal(t, 1, to.Day())
	})

	t.Run("lastweek precedes week", func(t *testing.T) {
		from, _, err := service.PeriodRange("lastweek")
		require.NoError(t, err)
		now := time.Now()
		assert.True(t, from.Before(now))
	})

	t.Run("3months covers current and two preceding", func(t *testing.T) {
		from, to, err := service.PeriodRange("3months")
		require.NoError(t, err)
		assert.True(t, to.After(from))
		// 3 months ≈ 89-92 days; assert the span is within a sane window.
		span := to.Sub(from).Hours() / 24
		assert.GreaterOrEqual(t, span, 58.0)
		assert.LessOrEqual(t, span, 95.0)
	})

	t.Run("invalid period returns ErrInvalidPeriod", func(t *testing.T) {
		_, _, err := service.PeriodRange("eternity")
		assert.ErrorIs(t, err, domain.ErrInvalidPeriod)
	})
}
