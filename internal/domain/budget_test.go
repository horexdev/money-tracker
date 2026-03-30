package domain_test

import (
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestBudget_UsagePercent(t *testing.T) {
	tests := []struct {
		name       string
		limit      int64
		spent      int64
		wantApprox float64
	}{
		{"zero limit returns 0", 0, 100, 0},
		{"half spent", 10000, 5000, 50.0},
		{"fully spent", 10000, 10000, 100.0},
		{"over limit", 10000, 15000, 150.0},
		{"nothing spent", 10000, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{LimitCents: tt.limit, SpentCents: tt.spent}
			got := b.UsagePercent()
			assert.InDelta(t, tt.wantApprox, got, 0.01)
		})
	}
}

func TestBudget_IsOverLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int64
		spent int64
		want  bool
	}{
		{"spent less than limit", 10000, 9999, false},
		{"spent equals limit (boundary)", 10000, 10000, true},
		{"spent over limit", 10000, 10001, true},
		{"zero spent", 10000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{LimitCents: tt.limit, SpentCents: tt.spent}
			assert.Equal(t, tt.want, b.IsOverLimit())
		})
	}
}

func TestBudget_ShouldNotify(t *testing.T) {
	tests := []struct {
		name            string
		limit           int64
		spent           int64
		notifyAtPercent int
		want            bool
	}{
		{"usage below threshold", 10000, 5000, 80, false},
		{"usage exactly at threshold (boundary)", 10000, 8000, 80, true},
		{"usage above threshold", 10000, 9000, 80, true},
		{"zero limit always returns false", 0, 100, 80, false},
		{"threshold 0 always notifies", 10000, 1, 0, true},
		{"fully spent above 100% threshold", 10000, 10000, 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &domain.Budget{
				LimitCents:      tt.limit,
				SpentCents:      tt.spent,
				NotifyAtPercent: tt.notifyAtPercent,
			}
			assert.Equal(t, tt.want, b.ShouldNotify())
		})
	}
}
