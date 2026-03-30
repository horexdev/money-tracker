package domain_test

import (
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestSavingsGoal_ProgressPercent(t *testing.T) {
	tests := []struct {
		name    string
		target  int64
		current int64
		want    float64
	}{
		{"zero target returns 0", 0, 100, 0},
		{"half funded", 10000, 5000, 50.0},
		{"exactly funded", 10000, 10000, 100.0},
		{"overfunded capped at 100", 10000, 15000, 100.0},
		{"nothing saved", 10000, 0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{TargetCents: tt.target, CurrentCents: tt.current}
			assert.InDelta(t, tt.want, g.ProgressPercent(), 0.01)
		})
	}
}

func TestSavingsGoal_IsCompleted(t *testing.T) {
	tests := []struct {
		name    string
		target  int64
		current int64
		want    bool
	}{
		{"current less than target", 10000, 9999, false},
		{"current equals target", 10000, 10000, true},
		{"overfunded", 10000, 10001, true},
		{"nothing saved", 10000, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{TargetCents: tt.target, CurrentCents: tt.current}
			assert.Equal(t, tt.want, g.IsCompleted())
		})
	}
}

func TestSavingsGoal_RemainingCents(t *testing.T) {
	tests := []struct {
		name    string
		target  int64
		current int64
		want    int64
	}{
		{"partial progress", 10000, 3000, 7000},
		{"fully funded returns 0", 10000, 10000, 0},
		{"overfunded returns 0 not negative", 10000, 15000, 0},
		{"nothing saved", 10000, 0, 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &domain.SavingsGoal{TargetCents: tt.target, CurrentCents: tt.current}
			assert.Equal(t, tt.want, g.RemainingCents())
		})
	}
}
