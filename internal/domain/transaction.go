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
// ExchangeRateSnapshot is the rate from CurrencyCode to BaseCurrencyAtCreation at creation time.
type Transaction struct {
	ID                     int64
	UserID                 int64
	Type                   TransactionType
	AmountCents            int64
	CategoryID             int64
	CategoryName           string
	CategoryEmoji          string
	CategoryColor          string
	Note                   string
	CurrencyCode           string
	ExchangeRateSnapshot   float64
	BaseCurrencyAtCreation string
	AccountID              int64  // 0 means no account assigned (legacy)
	AccountName            string // joined for display
	CreatedAt              time.Time
	IsAdjustment           bool   // true = hidden from history/stats; affects balance only
}

// BalanceByCurrency holds income and expense totals for a single currency.
type BalanceByCurrency struct {
	CurrencyCode string
	IncomeCents  int64
	ExpenseCents int64
}

// CategoryStat aggregates spending or income by category for a given period.
type CategoryStat struct {
	CategoryName  string
	CategoryEmoji string
	CategoryColor string
	Type          TransactionType
	TotalCents    int64
	TxCount       int64
	CurrencyCode  string
}
