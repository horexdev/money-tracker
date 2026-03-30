package domain

import "time"

// Transfer represents a movement of funds between two accounts owned by the same user.
// Transfers are not counted as income or expense in stats.
type Transfer struct {
	ID               int64
	UserID           int64
	FromAccountID    int64
	ToAccountID      int64
	FromAccountName  string // joined for display
	ToAccountName    string // joined for display
	AmountCents      int64
	FromCurrencyCode string
	ToCurrencyCode   string
	ExchangeRate     float64
	Note             string
	CreatedAt        time.Time
	// FromTxID and ToTxID are the IDs of the auto-created debit/credit transactions.
	// Nil for transfers created before migration 00020.
	FromTxID *int64
	ToTxID   *int64
}
