package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/repository/sqlcgen"
)

// ExchangeSnapshotRepository stores and retrieves historical exchange rates.
type ExchangeSnapshotRepository struct {
	q *sqlcgen.Queries
}

// NewExchangeSnapshotRepository creates a new repository backed by the connection pool.
func NewExchangeSnapshotRepository(pool *pgxpool.Pool) *ExchangeSnapshotRepository {
	return &ExchangeSnapshotRepository{q: sqlcgen.New(pool)}
}

// Upsert inserts or updates a single exchange rate snapshot.
func (r *ExchangeSnapshotRepository) Upsert(ctx context.Context, date time.Time, base, target string, rate float64) error {
	return r.q.UpsertExchangeRate(ctx, sqlcgen.UpsertExchangeRateParams{
		SnapshotDate:   pgtype.Date{Time: date, Valid: true},
		BaseCurrency:   base,
		TargetCurrency: target,
		Rate:           pgNumeric(rate),
	})
}

// GetRate returns the exchange rate for a specific date and currency pair.
// Returns 0, ErrNoRows if no rate exists.
func (r *ExchangeSnapshotRepository) GetRate(ctx context.Context, date time.Time, base, target string) (float64, error) {
	rate, err := r.q.GetExchangeRate(ctx, sqlcgen.GetExchangeRateParams{
		SnapshotDate:   pgtype.Date{Time: date, Valid: true},
		BaseCurrency:   base,
		TargetCurrency: target,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("no rate for %s->%s on %s: %w", base, target, date.Format("2006-01-02"), pgx.ErrNoRows)
		}
		return 0, err
	}
	return goFloat64(rate), nil
}

// GetRateOrLatest returns the rate for the given date, or the most recent rate before it.
func (r *ExchangeSnapshotRepository) GetRateOrLatest(ctx context.Context, date time.Time, base, target string) (float64, error) {
	rate, err := r.q.GetExchangeRateOrLatest(ctx, sqlcgen.GetExchangeRateOrLatestParams{
		BaseCurrency:   base,
		TargetCurrency: target,
		SnapshotDate:   pgtype.Date{Time: date, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("no rate for %s->%s on or before %s: %w", base, target, date.Format("2006-01-02"), pgx.ErrNoRows)
		}
		return 0, err
	}
	return goFloat64(rate), nil
}

// ListDistinctBaseCurrencies returns all distinct currency codes from accounts.
func (r *ExchangeSnapshotRepository) ListDistinctBaseCurrencies(ctx context.Context) ([]string, error) {
	return r.q.ListDistinctBaseCurrencies(ctx)
}

// GetLatestSnapshotDate returns the most recent snapshot date, or epoch if none.
func (r *ExchangeSnapshotRepository) GetLatestSnapshotDate(ctx context.Context) (time.Time, error) {
	d, err := r.q.GetLatestSnapshotDate(ctx)
	if err != nil {
		return time.Time{}, err
	}
	if !d.Valid {
		return time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), nil
	}
	return d.Time, nil
}
