package domain

// Category is a label for grouping transactions.
// System categories have UserID == 0 (NULL in DB).
type Category struct {
	ID     int64
	UserID int64
	Name   string
	Emoji  string
}

// IsSystem returns true for built-in categories shared by all users.
func (c *Category) IsSystem() bool {
	return c.UserID == 0
}
