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
type Transaction struct {
	ID            int64
	UserID        int64
	Type          TransactionType
	AmountCents   int64
	CategoryID    int64
	CategoryName  string
	CategoryEmoji string
	Note          string
	CreatedAt     time.Time
}

// CategoryStat aggregates spending or income by category for a given period.
type CategoryStat struct {
	CategoryName  string
	CategoryEmoji string
	Type          TransactionType
	TotalCents    int64
	TxCount       int64
}
