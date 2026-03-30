package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
)

// AdminStats holds aggregated metrics for the admin dashboard.
type AdminStats struct {
	TotalUsers     int64   `json:"total_users"`
	NewToday       int64   `json:"new_today"`
	NewThisWeek    int64   `json:"new_this_week"`
	NewThisMonth   int64   `json:"new_this_month"`
	RetentionDay1  float64 `json:"retention_day1"`
	RetentionDay7  float64 `json:"retention_day7"`
	RetentionDay30 float64 `json:"retention_day30"`
}

// AdminService provides admin-only analytics and user management.
type AdminService struct {
	repo AdminStorer
	log  *slog.Logger
}

func NewAdminService(repo AdminStorer, log *slog.Logger) *AdminService {
	return &AdminService{repo: repo, log: log}
}

// ListUsers returns a paginated user list and the total count.
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	total, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	offset := (page - 1) * pageSize
	users, err := s.repo.ListUsers(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

// GetStats computes aggregated admin metrics.
func (s *AdminService) GetStats(ctx context.Context) (*AdminStats, error) {
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	weekStart := todayStart.AddDate(0, 0, -int(todayStart.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	total, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	newToday, err := s.repo.CountNewUsers(ctx, todayStart, now)
	if err != nil {
		return nil, fmt.Errorf("count new users today: %w", err)
	}

	newWeek, err := s.repo.CountNewUsers(ctx, weekStart, now)
	if err != nil {
		return nil, fmt.Errorf("count new users this week: %w", err)
	}

	newMonth, err := s.repo.CountNewUsers(ctx, monthStart, now)
	if err != nil {
		return nil, fmt.Errorf("count new users this month: %w", err)
	}

	retDay1 := s.computeRetention(ctx, now, 1)
	retDay7 := s.computeRetention(ctx, now, 7)
	retDay30 := s.computeRetention(ctx, now, 30)

	return &AdminStats{
		TotalUsers:     total,
		NewToday:       newToday,
		NewThisWeek:    newWeek,
		NewThisMonth:   newMonth,
		RetentionDay1:  retDay1,
		RetentionDay7:  retDay7,
		RetentionDay30: retDay30,
	}, nil
}

// computeRetention calculates the retention rate for a cohort that signed up
// exactly `daysAgo` days ago. Retention = users who have any transaction after signup.
func (s *AdminService) computeRetention(ctx context.Context, now time.Time, daysAgo int) float64 {
	cohortStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -daysAgo)
	cohortEnd := cohortStart.AddDate(0, 0, 1)

	cohortSize, err := s.repo.CountNewUsers(ctx, cohortStart, cohortEnd)
	if err != nil || cohortSize == 0 {
		return 0
	}

	retained, err := s.repo.CountRetainedUsers(ctx, cohortStart, cohortEnd, cohortEnd, now)
	if err != nil {
		return 0
	}

	return float64(retained) / float64(cohortSize) * 100
}
