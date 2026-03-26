package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrCategoryNotFound    = errors.New("category not found")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidAmount       = errors.New("invalid amount: must be a positive number")
	ErrInvalidPeriod       = errors.New("invalid period")
)
