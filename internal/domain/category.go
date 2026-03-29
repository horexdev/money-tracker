package domain

import "time"

// CategoryType defines which transaction types a category applies to.
type CategoryType string

const (
	CategoryTypeExpense CategoryType = "expense"
	CategoryTypeIncome  CategoryType = "income"
	CategoryTypeBoth    CategoryType = "both"
)

// Category is a label for grouping transactions.
// System categories have UserID == 0 (NULL in DB).
type Category struct {
	ID        int64
	UserID    int64
	Name      string
	Emoji     string
	Type      CategoryType
	Color     string
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// IsSystem returns true for built-in categories shared by all users.
func (c *Category) IsSystem() bool {
	return c.UserID == 0
}
