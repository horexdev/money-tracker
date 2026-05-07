package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTranslateCategory_KnownLanguageAndKey(t *testing.T) {
	assert.Equal(t, "Еда", translateCategory("ru", "Food"))
	assert.Equal(t, "Транспорт", translateCategory("ru", "Transport"))
	assert.Equal(t, "Comida", translateCategory("es", "Food"))
}

func TestTranslateCategory_UnknownLanguageFallsBack(t *testing.T) {
	// Unknown lang returns the English key as-is.
	assert.Equal(t, "Food", translateCategory("xx", "Food"))
}

func TestTranslateCategory_UnknownKeyReturnsKey(t *testing.T) {
	// Known lang, unknown key — return key (English fallback).
	assert.Equal(t, "Custom", translateCategory("ru", "Custom"))
}

func TestAlertEmoji_BoundaryValues(t *testing.T) {
	cases := []struct {
		percent int
		want    string
	}{
		{0, "🟢"},
		{50, "🟢"},
		{74, "🟢"},
		{75, "🟡"},
		{94, "🟡"},
		{95, "🟠"},
		{99, "🟠"},
		{100, "🔴"},
		{200, "🔴"},
	}
	for _, c := range cases {
		assert.Equalf(t, c.want, alertEmoji(c.percent), "percent=%d", c.percent)
	}
}

func TestFormatMoney(t *testing.T) {
	assert.Equal(t, "10.00 USD", formatMoney(1000, "USD"))
	assert.Equal(t, "12.34 EUR", formatMoney(1234, "EUR"))
	assert.Equal(t, "0.05 USD", formatMoney(5, "USD"))
}

func TestFormatMoney_NegativeCents(t *testing.T) {
	// Negative cents render as a negative whole part with positive fractional digits.
	got := formatMoney(-1234, "USD")
	assert.Equal(t, "-12.34 USD", got)
}

func TestBudgetAlertTexts_AllLanguagesHaveFourEntries(t *testing.T) {
	for lang, labels := range budgetAlertTexts {
		assert.NotEmpty(t, labels[0], "lang %q title", lang)
		assert.NotEmpty(t, labels[1], "lang %q category label", lang)
		assert.NotEmpty(t, labels[2], "lang %q spent label", lang)
		assert.NotEmpty(t, labels[3], "lang %q remaining label", lang)
	}
}

func TestCategoryNames_EveryLocaleCoversBaselineKeys(t *testing.T) {
	en := categoryNames["en"]
	for lang, names := range categoryNames {
		for key := range en {
			_, ok := names[key]
			assert.Truef(t, ok, "lang %q missing translation for %q", lang, key)
		}
	}
}
