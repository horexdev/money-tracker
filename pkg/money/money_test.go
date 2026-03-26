package money_test

import (
	"testing"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/pkg/money"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCents(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr error
	}{
		{"12.50", 1250, nil},
		{"12", 1200, nil},
		{"0.01", 1, nil},
		{"100.00", 10000, nil},
		{"1.5", 150, nil},
		{"0", 0, domain.ErrInvalidAmount},
		{"0.00", 0, domain.ErrInvalidAmount},
		{"-5", 0, domain.ErrInvalidAmount},
		{"abc", 0, domain.ErrInvalidAmount},
		{"12.999", 0, domain.ErrInvalidAmount},
		{"", 0, domain.ErrInvalidAmount},
		{"12.5.3", 0, domain.ErrInvalidAmount},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := money.ParseCents(tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatCents(t *testing.T) {
	tests := []struct {
		cents int64
		want  string
	}{
		{1250, "12.50"},
		{1200, "12.00"},
		{1, "0.01"},
		{100, "1.00"},
		{0, "0.00"},
		{-1250, "-12.50"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := money.FormatCents(tt.cents)
			assert.Equal(t, tt.want, got)
		})
	}
}
