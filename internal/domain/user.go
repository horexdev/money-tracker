package domain

import "time"

// Language represents supported UI languages.
type Language string

const (
	LangEN Language = "en"
	LangRU Language = "ru"
)

// ValidLanguage checks if the given string is a supported language code.
func ValidLanguage(code string) bool {
	return code == string(LangEN) || code == string(LangRU)
}

// User represents a registered Telegram user.
type User struct {
	ID                int64
	Username          string
	FirstName         string
	LastName          string
	CurrencyCode      string
	Language          Language
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
