package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/horexdev/money-tracker/internal/domain"
)

// ExchangeService provides currency conversion using cached exchange rates.
type ExchangeService struct {
	provider RateProvider
	rdb      *redis.Client
	ttl      time.Duration
	log      *slog.Logger
}

// NewExchangeService creates a new exchange service with the given cache TTL.
func NewExchangeService(provider RateProvider, rdb *redis.Client, ttl time.Duration, log *slog.Logger) *ExchangeService {
	return &ExchangeService{
		provider: provider,
		rdb:      rdb,
		ttl:      ttl,
		log:      log,
	}
}

// ratesCacheKey returns the Redis key for cached rates.
func ratesCacheKey(base string) string {
	return fmt.Sprintf("exchange:%s", base)
}

// GetRate returns the exchange rate from one currency to another.
func (s *ExchangeService) GetRate(ctx context.Context, from, to string) (float64, error) {
	if from == to {
		return 1.0, nil
	}

	rates, err := s.getRates(ctx, from)
	if err != nil {
		return 0, err
	}

	rate, ok := rates[to]
	if !ok {
		return 0, fmt.Errorf("rate %s->%s: %w", from, to, domain.ErrExchangeRateUnavailable)
	}
	return rate, nil
}

// Convert converts an amount in cents from one currency to another.
func (s *ExchangeService) Convert(ctx context.Context, amountCents int64, from, to string) (int64, error) {
	rate, err := s.GetRate(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return int64(math.Round(float64(amountCents) * rate)), nil
}

// ConvertMulti converts an amount in cents from one currency to multiple target currencies.
func (s *ExchangeService) ConvertMulti(ctx context.Context, amountCents int64, from string, to []string) (map[string]int64, error) {
	if len(to) == 0 {
		return nil, nil
	}

	rates, err := s.getRates(ctx, from)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(to))
	for _, currency := range to {
		if currency == from {
			result[currency] = amountCents
			continue
		}
		rate, ok := rates[currency]
		if !ok {
			continue // skip unavailable rates silently
		}
		result[currency] = int64(math.Round(float64(amountCents) * rate))
	}
	return result, nil
}

// getRates returns exchange rates for a base currency, using Redis cache with fallback to API.
func (s *ExchangeService) getRates(ctx context.Context, base string) (map[string]float64, error) {
	key := ratesCacheKey(base)

	// Try cache first.
	cached, err := s.rdb.Get(ctx, key).Result()
	if err == nil {
		var rates map[string]float64
		if err := json.Unmarshal([]byte(cached), &rates); err == nil {
			return rates, nil
		}
	}

	// Cache miss or decode error — fetch from API.
	rates, err := s.provider.FetchRates(ctx, base)
	if err != nil {
		s.log.WarnContext(ctx, "exchange rate fetch failed", slog.String("base", base), slog.String("error", err.Error()))
		return nil, fmt.Errorf("fetch rates for %s: %w", base, domain.ErrExchangeRateUnavailable)
	}

	// Cache with TTL.
	data, err := json.Marshal(rates)
	if err == nil {
		if err := s.rdb.Set(ctx, key, data, s.ttl).Err(); err != nil {
			s.log.WarnContext(ctx, "failed to cache exchange rates", slog.String("error", err.Error()))
		}
	}

	return rates, nil
}
