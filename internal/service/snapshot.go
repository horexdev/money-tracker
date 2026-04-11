package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// SnapshotStorer is the repository interface for exchange rate snapshots.
type SnapshotStorer interface {
	Upsert(ctx context.Context, date time.Time, base, target string, rate float64) error
	GetRate(ctx context.Context, date time.Time, base, target string) (float64, error)
	GetRateOrLatest(ctx context.Context, date time.Time, base, target string) (float64, error)
	ListDistinctBaseCurrencies(ctx context.Context) ([]string, error)
	GetLatestSnapshotDate(ctx context.Context) (time.Time, error)
}

// SnapshotService manages daily exchange rate snapshots.
type SnapshotService struct {
	repo     SnapshotStorer
	provider RateProvider
	log      *slog.Logger
}

// NewSnapshotService creates a new snapshot service.
func NewSnapshotService(repo SnapshotStorer, provider RateProvider, log *slog.Logger) *SnapshotService {
	return &SnapshotService{
		repo:     repo,
		provider: provider,
		log:      log,
	}
}

// SaveDaily fetches current exchange rates for all currencies used in accounts
// and stores them as today's snapshot. Safe to call multiple times per day (upsert).
func (s *SnapshotService) SaveDaily(ctx context.Context) error {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	bases, err := s.repo.ListDistinctBaseCurrencies(ctx)
	if err != nil {
		return fmt.Errorf("list base currencies: %w", err)
	}
	if len(bases) == 0 {
		s.log.InfoContext(ctx, "snapshot: no accounts found, skipping")
		return nil
	}

	var totalSaved int
	for _, base := range bases {
		rates, err := s.provider.FetchRates(ctx, base)
		if err != nil {
			s.log.ErrorContext(ctx, "snapshot: failed to fetch rates",
				slog.String("base", base),
				slog.String("error", err.Error()),
			)
			continue
		}

		for target, rate := range rates {
			if target == base {
				continue
			}
			if err := s.repo.Upsert(ctx, today, base, target, rate); err != nil {
				s.log.ErrorContext(ctx, "snapshot: failed to upsert rate",
					slog.String("base", base),
					slog.String("target", target),
					slog.String("error", err.Error()),
				)
				continue
			}
			totalSaved++
		}
	}

	s.log.InfoContext(ctx, "snapshot: daily rates saved",
		slog.Int("bases", len(bases)),
		slog.Int("rates_saved", totalSaved),
		slog.String("date", today.Format("2006-01-02")),
	)
	return nil
}

// GetRate returns the exchange rate for a specific date and currency pair.
// Falls back to the most recent available rate if the exact date is missing.
func (s *SnapshotService) GetRate(ctx context.Context, date time.Time, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	rate, err := s.repo.GetRateOrLatest(ctx, date, from, to)
	if err != nil {
		return 0, fmt.Errorf("snapshot rate %s->%s on %s: %w", from, to, date.Format("2006-01-02"), domain.ErrExchangeRateUnavailable)
	}
	return rate, nil
}

// Convert converts an amount in cents from one currency to another using the
// historical rate for the given date.
func (s *SnapshotService) Convert(ctx context.Context, amountCents int64, from, to string, date time.Time) (int64, error) {
	rate, err := s.GetRate(ctx, date, from, to)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(float64(amountCents) * rate)), nil
}
