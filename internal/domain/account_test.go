package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/horexdev/money-tracker/internal/domain"
)

// TestAccountType_StringValues guards against accidental edits to the AccountType
// constants. The string values are part of the database enum (account_type) and
// must match exactly — a mismatch would silently break inserts.
func TestAccountType_StringValues(t *testing.T) {
	cases := map[domain.AccountType]string{
		domain.AccountTypeChecking: "checking",
		domain.AccountTypeSavings:  "savings",
		domain.AccountTypeCash:     "cash",
		domain.AccountTypeCredit:   "credit",
		domain.AccountTypeCrypto:   "crypto",
	}
	for got, want := range cases {
		assert.Equal(t, want, string(got), "AccountType %q must serialise as %q", got, want)
	}
}
