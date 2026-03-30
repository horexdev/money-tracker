// Package api exports internal symbols for use in api_test package tests.
package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/service"
)

// HttpStatus exposes httpStatus for testing.
var HttpStatus = httpStatus

// UserMessage exposes userMessage for testing.
var UserMessage = userMessage

// WithUserID returns a context with the given user ID injected.
// Used by handler tests to simulate authenticated requests.
func WithUserID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, contextKeyUserID, id)
}

// UserIDFromContextForTest exposes userIDFromContext for testing.
func UserIDFromContextForTest(ctx context.Context) int64 {
	return userIDFromContext(ctx)
}

// testEnsureAdapter adapts an external EnsureUser func to the internal ensurer interface.
type testEnsureAdapter struct {
	fn func(ctx context.Context, u TelegramUser) error
}

func (t *testEnsureAdapter) ensureUser(ctx context.Context, u TelegramUser) error {
	return t.fn(ctx, u)
}

// EnsureUserFunc is the function signature for ensureUser, exposed for tests.
type EnsureUserFunc func(ctx context.Context, u TelegramUser) error

// AuthMiddlewareForTest exposes authMiddleware for testing using an EnsureUserFunc.
func AuthMiddlewareForTest(botToken string, devMode bool, fn EnsureUserFunc) func(http.Handler) http.Handler {
	return authMiddleware(botToken, devMode, "en", &testEnsureAdapter{fn: fn}, slog.Default())
}

// AdminMiddlewareForTest exposes adminMiddleware for testing.
func AdminMiddlewareForTest(adminUserID int64) func(http.Handler) http.Handler {
	return adminMiddleware(adminUserID, false)
}

// CorsMiddlewareForTest exposes corsMiddleware for testing.
func CorsMiddlewareForTest(allowedOrigins string) func(http.Handler) http.Handler {
	return corsMiddleware(allowedOrigins)
}

// TransactionHandlerForTest exposes transactionHandler for testing.
// The exchangeSvc is nil since handler tests use type="expense"/"income" without currency conversion
// (userSvc.GetByID returns same currency as request so no exchange rate lookup is needed).
func TransactionHandlerForTest(txSvc *service.TransactionService, userSvc *service.UserService, log *slog.Logger) http.HandlerFunc {
	return transactionHandler(txSvc, userSvc, nil, log)
}

// SettingsHandlerForTest exposes settingsHandler for testing.
func SettingsHandlerForTest(userSvc *service.UserService, adminUserID int64, log *slog.Logger) http.HandlerFunc {
	return settingsHandler(userSvc, adminUserID, false, log)
}

// CategoriesHandlerForTest exposes categoriesHandler for testing.
func CategoriesHandlerForTest(catSvc *service.CategoryService, log *slog.Logger) http.HandlerFunc {
	return categoriesHandler(catSvc, log)
}

// AccountsHandlerForTest exposes accountsHandler for testing.
func AccountsHandlerForTest(accountSvc *service.AccountService, log *slog.Logger) http.HandlerFunc {
	return accountsHandler(accountSvc, log)
}

// BudgetHandlerForTest exposes budgetHandler for testing.
func BudgetHandlerForTest(budgetSvc *service.BudgetService, log *slog.Logger) http.HandlerFunc {
	return budgetHandler(budgetSvc, log)
}

// GoalsHandlerForTest exposes goalsHandler for testing.
func GoalsHandlerForTest(goalSvc *service.SavingsGoalService, log *slog.Logger) http.HandlerFunc {
	return goalsHandler(goalSvc, log)
}

// RecurringHandlerForTest exposes recurringHandler for testing.
func RecurringHandlerForTest(recurringSvc *service.RecurringService, log *slog.Logger) http.HandlerFunc {
	return recurringHandler(recurringSvc, log)
}

// TransfersHandlerForTest exposes transfersHandler for testing.
func TransfersHandlerForTest(transferSvc *service.TransferService, log *slog.Logger) http.HandlerFunc {
	return transfersHandler(transferSvc, log)
}

// AdminUsersHandlerForTest exposes adminUsersHandler for testing.
func AdminUsersHandlerForTest(adminSvc *service.AdminService, log *slog.Logger) http.HandlerFunc {
	return adminUsersHandler(adminSvc, log)
}

// TelegramUserForTest constructs a TelegramUser for use in tests.
func TelegramUserForTest(id int64, firstName, langCode string) TelegramUser {
	return TelegramUser{ID: id, FirstName: firstName, LanguageCode: langCode}
}

// LocalizedAccountName exposes localizedAccountName for testing.
func LocalizedAccountName(lang string) string { return localizedAccountName(lang) }

// LocalizedAccountCurrency exposes localizedAccountCurrency for testing.
func LocalizedAccountCurrency(lang string) string { return localizedAccountCurrency(lang) }

// NewUserEnsurer constructs a userEnsurer for testing.
func NewUserEnsurer(userSvc *service.UserService, accountSvc *service.AccountService, log *slog.Logger) *userEnsurer {
	return &userEnsurer{svc: userSvc, accountSvc: accountSvc, log: log}
}

// EnsureUserForTest calls ensureUser on a userEnsurer, exposing it for tests.
func EnsureUserForTest(ue *userEnsurer, ctx context.Context, tgUser TelegramUser) error {
	return ue.ensureUser(ctx, tgUser)
}
