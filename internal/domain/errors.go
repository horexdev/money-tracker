package domain

import "errors"

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrCategoryNotFound         = errors.New("category not found")
	ErrTransactionNotFound      = errors.New("transaction not found")
	ErrBudgetNotFound           = errors.New("budget not found")
	ErrRecurringNotFound        = errors.New("recurring transaction not found")
	ErrGoalNotFound             = errors.New("savings goal not found")
	ErrInvalidAmount            = errors.New("invalid amount: must be a positive number")
	ErrInvalidPeriod            = errors.New("invalid period")
	ErrInvalidCurrency          = errors.New("invalid currency code")
	ErrInvalidFrequency         = errors.New("invalid frequency")
	ErrInvalidLanguage          = errors.New("invalid language: must be en or ru")
	ErrTooManyDisplayCurrencies = errors.New("maximum 3 display currencies allowed")
	ErrExchangeRateUnavailable  = errors.New("exchange rate unavailable")
	ErrCategoryInUse            = errors.New("category is referenced by transactions")
	ErrCategorySystemReadOnly   = errors.New("system categories cannot be modified")
	ErrInsufficientGoalFunds    = errors.New("insufficient funds in savings goal")
	ErrBudgetAlreadyExists      = errors.New("budget already exists for this category and period")
	ErrAccountNotFound          = errors.New("account not found")
	ErrDefaultAccountExists     = errors.New("default account already exists")
	ErrAccountHasTransactions   = errors.New("account has transactions and cannot be deleted")
	ErrTransferSameAccount      = errors.New("transfer source and destination must be different")
)
