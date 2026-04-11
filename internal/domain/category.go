package domain

import "time"

// CategoryType defines which transaction types a category applies to.
type CategoryType string

const (
	CategoryTypeExpense    CategoryType = "expense"
	CategoryTypeIncome     CategoryType = "income"
	CategoryTypeBoth       CategoryType = "both"
	CategoryTypeSavings    CategoryType = "savings"
	CategoryTypeTransfer   CategoryType = "transfer"
	CategoryTypeAdjustment CategoryType = "adjustment"
)

// Category is a label for grouping transactions.
// Personal categories have UserID != 0.
// Infrastructure categories (Transfer, Adjustment) have UserID == 0 and IsProtected == true.
type Category struct {
	ID          int64
	UserID      int64
	Name        string
	Icon        string
	Type        CategoryType
	Color       string
	IsProtected bool
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// CategorySeed holds the data needed to create a default category for a user.
type CategorySeed struct {
	Name  string
	Icon  string
	Type  CategoryType
	Color string
}

// IsSystem returns true for infrastructure categories (Transfer, Adjustment).
// These are kept with user_id = NULL and cannot be modified by users.
func (c *Category) IsSystem() bool {
	return c.UserID == 0
}

// IsPersonal returns true for user-owned categories.
func (c *Category) IsPersonal() bool {
	return c.UserID != 0
}
