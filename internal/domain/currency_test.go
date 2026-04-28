package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/domain"
)

func TestValidCurrency(t *testing.T) {
	cases := []struct {
		code string
		want bool
	}{
		{"USD", true},
		{"EUR", true},
		{"GBP", true},
		{"usd", true},
		{"  USD  ", false}, // ToUpper only — does not trim
		{"XYZ", false},
		{"", false},
	}
	for _, c := range cases {
		t.Run(c.code, func(t *testing.T) {
			assert.Equal(t, c.want, domain.ValidCurrency(c.code))
		})
	}
}

func TestSearchCurrencies_EmptyQueryReturnsNil(t *testing.T) {
	assert.Nil(t, domain.SearchCurrencies(""))
}

func TestSearchCurrencies_CodePrefixMatch(t *testing.T) {
	results := domain.SearchCurrencies("USD")
	require.NotEmpty(t, results)
	assert.Equal(t, "USD", results[0].Code, "exact code prefix must rank first")
}

func TestSearchCurrencies_CapsAtFiveResults(t *testing.T) {
	results := domain.SearchCurrencies("E")
	assert.LessOrEqual(t, len(results), 5)
}

func TestSearchCurrencies_NameContainsMatch(t *testing.T) {
	results := domain.SearchCurrencies("Dollar")
	require.NotEmpty(t, results)
	found := false
	for _, c := range results {
		if strings.Contains(strings.ToLower(c.Name), "dollar") {
			found = true
			break
		}
	}
	assert.True(t, found, "results must contain at least one currency with 'dollar' in its name")
}

func TestSearchCurrencies_NoDuplicatesBetweenPasses(t *testing.T) {
	results := domain.SearchCurrencies("US")
	seen := make(map[string]bool)
	for _, c := range results {
		assert.Falsef(t, seen[c.Code], "duplicate currency in results: %s", c.Code)
		seen[c.Code] = true
	}
}
