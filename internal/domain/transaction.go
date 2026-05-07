package domain

import "time"

// TransactionType distinguishes expense from income entries.
type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
)

// Transaction is a single financial event recorded by the user.
// AmountCents stores the value in integer cents to avoid float precision issues.
// CurrencyCode is the account's currency. SnapshotDate identifies which exchange rate to use
// when converting to another currency (looked up from exchange_rate_snapshots table).
type Transaction struct {
	ID            int64
	UserID        int64
	Type          TransactionType
	AmountCents   int64
	CategoryID    int64
	CategoryName  string
	CategoryIcon  string
	CategoryColor string
	Note          string
	CurrencyCode  string
	AccountID     int64
	AccountName   string // joined for display
	SnapshotDate  time.Time
	CreatedAt     time.Time
	IsAdjustment  bool // true = hidden from history/stats; affects balance only
}

// BalanceByCurrency holds income and expense totals for a single currency.
type BalanceByCurrency struct {
	CurrencyCode string
	IncomeCents  int64
	ExpenseCents int64
}

// CategoryStat aggregates spending or income by category for a given period.
type CategoryStat struct {
	CategoryID    int64
	CategoryName  string
	CategoryIcon string
	CategoryColor string
	Type          TransactionType
	TotalCents    int64
	TxCount       int64
	CurrencyCode  string
}
