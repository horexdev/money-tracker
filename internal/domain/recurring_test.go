package domain_test

import (
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestRecurringTransaction_NextRunAfter(t *testing.T) {
	base := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		frequency domain.Frequency
		after     time.Time
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "daily adds 1 day",
			frequency: domain.FrequencyDaily,
			after:     base,
			wantYear:  2024, wantMonth: time.March, wantDay: 16,
		},
		{
			name:      "weekly adds 7 days",
			frequency: domain.FrequencyWeekly,
			after:     base,
			wantYear:  2024, wantMonth: time.March, wantDay: 22,
		},
		{
			name:      "monthly adds 1 calendar month",
			frequency: domain.FrequencyMonthly,
			after:     base,
			wantYear:  2024, wantMonth: time.April, wantDay: 15,
		},
		{
			name:      "yearly adds 1 calendar year",
			frequency: domain.FrequencyYearly,
			after:     base,
			wantYear:  2025, wantMonth: time.March, wantDay: 15,
		},
		{
			name:      "unknown frequency falls back to monthly",
			frequency: domain.Frequency("unknown"),
			after:     base,
			wantYear:  2024, wantMonth: time.April, wantDay: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &domain.RecurringTransaction{Frequency: tt.frequency}
			got := rt.NextRunAfter(tt.after)
			assert.Equal(t, tt.wantYear, got.Year())
			assert.Equal(t, tt.wantMonth, got.Month())
			assert.Equal(t, tt.wantDay, got.Day())
		})
	}
}

func TestRecurringTransaction_NextRunAfter_MonthEndEdgeCases(t *testing.T) {
	t.Run("Jan 31 + monthly overflows to March in non-leap year", func(t *testing.T) {
		rt := &domain.RecurringTransaction{Frequency: domain.FrequencyMonthly}
		jan31 := time.Date(2023, time.January, 31, 0, 0, 0, 0, time.UTC)
		got := rt.NextRunAfter(jan31)
		// Go's AddDate(0,1,0) on Jan 31 overflows: Feb has 28 days → March 3
		assert.Equal(t, time.March, got.Month())
		assert.Equal(t, 3, got.Day())
		assert.Equal(t, 2023, got.Year())
	})

	t.Run("Jan 31 + monthly overflows to March in leap year", func(t *testing.T) {
		rt := &domain.RecurringTransaction{Frequency: domain.FrequencyMonthly}
		jan31 := time.Date(2024, time.January, 31, 0, 0, 0, 0, time.UTC)
		got := rt.NextRunAfter(jan31)
		// 2024 is a leap year: Feb has 29 days → March 2
		assert.Equal(t, time.March, got.Month())
		assert.Equal(t, 2, got.Day())
		assert.Equal(t, 2024, got.Year())
	})

	t.Run("Dec 31 + yearly crosses year boundary", func(t *testing.T) {
		rt := &domain.RecurringTransaction{Frequency: domain.FrequencyYearly}
		dec31 := time.Date(2023, time.December, 31, 0, 0, 0, 0, time.UTC)
		got := rt.NextRunAfter(dec31)
		assert.Equal(t, 2024, got.Year())
		assert.Equal(t, time.December, got.Month())
		assert.Equal(t, 31, got.Day())
	})
}
