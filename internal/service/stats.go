package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/repository"
)

// StatsService provides aggregated financial statistics.
type StatsService struct {
	txRepo *repository.TransactionRepository
	log    *slog.Logger
}

func NewStatsService(txRepo *repository.TransactionRepository, log *slog.Logger) *StatsService {
	return &StatsService{txRepo: txRepo, log: log}
}

// ByCategory returns spending/income grouped by category for the given period.
func (s *StatsService) ByCategory(ctx context.Context, userID int64, from, to time.Time) ([]domain.CategoryStat, error) {
	stats, err := s.txRepo.StatsByCategory(ctx, userID, from, to)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to get stats by category",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, err
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
	case "lastmonth":
		start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start, end, nil
	default:
		return time.Time{}, time.Time{}, domain.ErrInvalidPeriod
	}
}
