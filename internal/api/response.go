package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error response, mapping domain sentinel errors to
// appropriate HTTP status codes.
func writeError(w http.ResponseWriter, log *slog.Logger, err error) {
	status := httpStatus(err)
	if status == http.StatusInternalServerError {
		log.Error("internal error", slog.String("error", err.Error()))
	}
	writeJSON(w, status, errorResponse{Error: userMessage(err)})
}

// httpStatus maps domain errors to HTTP status codes.
func httpStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrCategoryNotFound),
		errors.Is(err, domain.ErrTransactionNotFound),
		errors.Is(err, domain.ErrBudgetNotFound),
		errors.Is(err, domain.ErrRecurringNotFound),
		errors.Is(err, domain.ErrGoalNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidPeriod),
		errors.Is(err, domain.ErrInvalidCurrency),
		errors.Is(err, domain.ErrInvalidFrequency),
		errors.Is(err, domain.ErrInvalidLanguage),
		errors.Is(err, domain.ErrTooManyDisplayCurrencies):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrBudgetAlreadyExists),
		errors.Is(err, domain.ErrCategoryInUse):
		return http.StatusConflict
	case errors.Is(err, domain.ErrCategorySystemReadOnly):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrInsufficientGoalFunds):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrExchangeRateUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// userMessage returns a safe, user-facing error message.
func userMessage(err error) string {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return "user not found"
	case errors.Is(err, domain.ErrCategoryNotFound):
		return "category not found"
	case errors.Is(err, domain.ErrTransactionNotFound):
		return "transaction not found"
	case errors.Is(err, domain.ErrBudgetNotFound):
		return "budget not found"
	case errors.Is(err, domain.ErrRecurringNotFound):
		return "recurring transaction not found"
	case errors.Is(err, domain.ErrGoalNotFound):
		return "savings goal not found"
	case errors.Is(err, domain.ErrInvalidAmount):
		return "invalid amount: must be a positive integer number of cents"
	case errors.Is(err, domain.ErrInvalidPeriod):
		return "invalid period: use today, week, month, lastweek, lastmonth, or 3months"
	case errors.Is(err, domain.ErrInvalidCurrency):
		return "invalid currency code"
	case errors.Is(err, domain.ErrInvalidFrequency):
		return "invalid frequency: use daily, weekly, monthly, or yearly"
	case errors.Is(err, domain.ErrInvalidLanguage):
		return "invalid language: use en or ru"
	case errors.Is(err, domain.ErrTooManyDisplayCurrencies):
		return "maximum 3 display currencies allowed"
	case errors.Is(err, domain.ErrBudgetAlreadyExists):
		return "budget already exists for this category and period"
	case errors.Is(err, domain.ErrCategoryInUse):
		return "category has transactions and cannot be deleted"
	case errors.Is(err, domain.ErrCategorySystemReadOnly):
		return "system categories cannot be modified"
	case errors.Is(err, domain.ErrInsufficientGoalFunds):
		return "insufficient funds in savings goal"
	case errors.Is(err, domain.ErrExchangeRateUnavailable):
		return "exchange rate temporarily unavailable"
	default:
		return "internal server error"
	}
}
