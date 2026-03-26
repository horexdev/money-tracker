// Package money provides decimal-safe helpers for monetary amounts.
// All values are stored as integer cents (e.g. $12.50 → 1250) to avoid
// floating-point precision issues.
package money

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/horexdev/money-tracker/internal/domain"
)

// ParseCents converts a human-typed string like "12.50" or "12" into
// integer cents. Returns domain.ErrInvalidAmount if the input is not a
// valid positive decimal with at most 2 decimal places.
func ParseCents(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, domain.ErrInvalidAmount
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, domain.ErrInvalidAmount
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || whole < 0 {
		return 0, domain.ErrInvalidAmount
	}

	var frac int64
	if len(parts) == 2 {
		f := parts[1]
		if len(f) == 0 || len(f) > 2 {
			return 0, domain.ErrInvalidAmount
		}
		if len(f) == 1 {
			f += "0"
		}
		frac, err = strconv.ParseInt(f, 10, 64)
		if err != nil {
			return 0, domain.ErrInvalidAmount
		}
	}

	cents := whole*100 + frac
	if cents <= 0 {
		return 0, domain.ErrInvalidAmount
	}
	return cents, nil
}

// FormatCents converts integer cents into a human-readable string like "12.50".
func FormatCents(cents int64) string {
	if cents < 0 {
		return fmt.Sprintf("-%s", FormatCents(-cents))
	}
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// CurrencySymbol returns the symbol for a given ISO 4217 currency code.
// Falls back to the code itself if unknown.
func CurrencySymbol(code string) string {
	if c, ok := domain.Currencies[code]; ok {
		return c.Symbol
	}
	return code
}
