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
func TransactionHandlerForTest(txSvc *service.TransactionService, accountSvc *service.AccountService, log *slog.Logger) http.HandlerFunc {
	return transactionHandler(txSvc, accountSvc, log)
}

// SettingsHandlerForTest exposes settingsHandler for testing.
func SettingsHandlerForTest(userSvc *service.UserService, accountSvc *service.AccountService, adminUserID int64, log *slog.Logger) http.HandlerFunc {
	return settingsHandler(userSvc, accountSvc, adminUserID, false, log)
}

// CategoriesHandlerForTest exposes categoriesHandler for testing.
func CategoriesHandlerForTest(catSvc *service.CategoryService, log *slog.Logger) http.HandlerFunc {
	return categoriesHandler(catSvc, log)
}

// AccountsHandlerForTest exposes accountsHandler for testing.
func AccountsHandlerForTest(accountSvc *service.AccountService, adjustSvc *service.AdjustmentService, log *slog.Logger) http.HandlerFunc {
	return accountsHandler(accountSvc, adjustSvc, log)
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

// BalanceFetcher is the exported alias for the unexported balanceFetcher
// interface, allowing api_test handlers to inject mock implementations.
type BalanceFetcher = balanceFetcher

// BalanceHandlerForTest exposes balanceHandler for testing. exchangeSvc may be
// nil if the user's DisplayCurrencies is empty.
func BalanceHandlerForTest(txSvc BalanceFetcher, userSvc *service.UserService, accountSvc *service.AccountService, exchangeSvc *service.ExchangeService, log *slog.Logger) http.HandlerFunc {
	return balanceHandler(txSvc, userSvc, accountSvc, exchangeSvc, log)
}

// StatsHandlerForTest exposes statsHandler for testing.
func StatsHandlerForTest(statsSvc *service.StatsService, log *slog.Logger) http.HandlerFunc {
	return statsHandler(statsSvc, log)
}

// AdjustmentApplier is the exported alias for the unexported adjustmentApplier
// interface used by adjustAccountHandler.
type AdjustmentApplier = adjustmentApplier

// AdjustAccountHandlerForTest exposes adjustAccountHandler for testing.
func AdjustAccountHandlerForTest(svc AdjustmentApplier, log *slog.Logger, accountID int64) http.HandlerFunc {
	return adjustAccountHandler(svc, log, accountID)
}

// UserDataHandlerForTest exposes userDataHandler for testing.
func UserDataHandlerForTest(userSvc *service.UserService, log *slog.Logger) http.HandlerFunc {
	return userDataHandler(userSvc, log)
}

// DefaultCategoriesFor exposes defaultCategoriesFor for tests.
var DefaultCategoriesFor = defaultCategoriesFor

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
	return &userEnsurer{svc: userSvc, accountSvc: accountSvc, categorySvc: nil, log: log}
}

// EnsureUserForTest calls ensureUser on a userEnsurer, exposing it for tests.
func EnsureUserForTest(ue *userEnsurer, ctx context.Context, tgUser TelegramUser) error {
	return ue.ensureUser(ctx, tgUser)
}
