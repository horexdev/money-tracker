package domain

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrCategoryNotFound          = errors.New("category not found")
	ErrTransactionNotFound       = errors.New("transaction not found")
	ErrInvalidAmount             = errors.New("invalid amount: must be a positive number")
	ErrInvalidPeriod             = errors.New("invalid period")
	ErrInvalidCurrency           = errors.New("invalid currency code")
	ErrTooManyDisplayCurrencies  = errors.New("maximum 3 display currencies allowed")
	ErrExchangeRateUnavailable   = errors.New("exchange rate unavailable")
)
