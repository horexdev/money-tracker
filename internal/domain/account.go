package domain

import "time"

// AccountType represents the kind of financial account.
type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCash     AccountType = "cash"
	AccountTypeCredit   AccountType = "credit"
	AccountTypeCrypto   AccountType = "crypto"
)

// Account is a named financial wallet/account owned by a user.
type Account struct {
	ID             int64
	UserID         int64
	Name           string
	Icon           string      // Phosphor icon ID, e.g. "credit-card"
	Color          string      // hex colour, e.g. "#6366f1"
	Type           AccountType
	CurrencyCode   string
	IsDefault      bool
	IncludeInTotal bool
	BalanceCents   int64 // derived from transactions, not stored
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
