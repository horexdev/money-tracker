package domain

import "time"

// User represents a registered Telegram user.
type User struct {
	ID                int64
	Username          string
	FirstName         string
	LastName          string
	CurrencyCode      string
	DisplayCurrencies []string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// DisplayName returns the best available name for the user.
func (u *User) DisplayName() string {
	if u.FirstName != "" {
		return u.FirstName
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	return "User"
}
