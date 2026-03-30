package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// StatsService provides aggregated financial statistics.
type StatsService struct {
	txRepo TransactionStorer
	log    *slog.Logger
}

func NewStatsService(txRepo TransactionStorer, log *slog.Logger) *StatsService {
	return &StatsService{txRepo: txRepo, log: log}
}

// ByCategory returns spending/income grouped by category for the given period.
func (s *StatsService) ByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	stats, err := s.txRepo.StatsByCategory(ctx, userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats by category for user %d: %w", userID, err)
	}
	return stats, nil
}

// ByCategoryAndAccount returns stats for a specific account.
func (s *StatsService) ByCategoryAndAccount(ctx context.Context, userID, accountID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	stats, err := s.txRepo.StatsByCategoryAndAccount(ctx, userID, accountID, from, to)
	if err != nil {
		return nil, fmt.Errorf("stats by category for account %d: %w", accountID, err)
	}
	return stats, nil
}

// PeriodRange returns the start and end time for a named period relative to now.
// Supported: "today", "week", "month", "lastmonth".
func PeriodRange(period string) (from, to time.Time, err error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch period {
	case "today":
		return today, today.Add(24 * time.Hour), nil
	case "week":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		return monday, monday.AddDate(0, 0, 7), nil
	case "month":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end := start.AddDate(0, 1, 0)
		return start, end, nil
	case "lastweek":
		weekday := int(today.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		monday := today.AddDate(0, 0, -(weekday - 1))
		start := monday.AddDate(0, 0, -7)
		return start, monday, nil
	case "lastmonth":
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		start := firstOfMonth.AddDate(0, -1, 0)
		end := firstOfMonth
		return start, end, nil
	case "3months":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -2, 0)
		end := today.Add(24 * time.Hour)
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, domain.ErrInvalidPeriod
	}
}
