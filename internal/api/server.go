package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

// userEnsurer wraps UserService to satisfy the ensurer interface required by authMiddleware.
type userEnsurer struct {
	svc *service.UserService
}

func (u *userEnsurer) ensureUser(ctx context.Context, userID int64) error {
	_, err := u.svc.Upsert(ctx, &domain.User{ID: userID})
	return err
}

// Deps holds all service dependencies needed to build the HTTP server.
type Deps struct {
	UserSvc        *service.UserService
	TxSvc          *service.TransactionService
	StatsSvc       *service.StatsService
	ExchangeSvc    *service.ExchangeService
	CategorySvc    *service.CategoryService
	BudgetSvc      *service.BudgetService
	RecurringSvc   *service.RecurringService
	GoalSvc        *service.SavingsGoalService
	ExportSvc      *service.ExportService
	BotToken       string
	AllowedOrigins string
	Log            *slog.Logger
}

// NewServer builds and returns the HTTP handler for the API server.
func NewServer(d Deps) http.Handler {
	mux := http.NewServeMux()

	eu := &userEnsurer{svc: d.UserSvc}

	auth := authMiddleware(d.BotToken, eu, d.Log)
	cors := corsMiddleware(d.AllowedOrigins)
	logging := loggingMiddleware(d.Log)

	// protected wraps a handler with logging + CORS + auth.
	protected := func(h http.HandlerFunc) http.Handler {
		return chain(h, logging, cors, auth)
	}

	// public wraps a handler with logging + CORS only.
	public := func(h http.HandlerFunc) http.Handler {
		return chain(h, logging, cors)
	}

	// Public endpoints (no Telegram auth required).
	mux.Handle("/api/v1/devblog", public(devblogListHandler(d.Log)))
	mux.Handle("/api/v1/devblog/", public(devblogEntryHandler(d.Log)))

	// Authenticated endpoints.
	mux.Handle("/api/v1/balance", protected(balanceHandler(d.TxSvc, d.UserSvc, d.ExchangeSvc, d.Log)))
	mux.Handle("/api/v1/transactions", protected(transactionHandler(d.TxSvc, d.Log)))
	mux.Handle("/api/v1/transactions/", protected(transactionHandler(d.TxSvc, d.Log)))
	mux.Handle("/api/v1/stats", protected(statsHandler(d.StatsSvc, d.Log)))
	mux.Handle("/api/v1/settings", protected(settingsHandler(d.UserSvc, d.Log)))

	// Categories CRUD.
	mux.Handle("/api/v1/categories", protected(categoriesHandler(d.CategorySvc, d.Log)))
	mux.Handle("/api/v1/categories/", protected(categoriesHandler(d.CategorySvc, d.Log)))

	// Budgets CRUD.
	mux.Handle("/api/v1/budgets", protected(budgetHandler(d.BudgetSvc, d.Log)))
	mux.Handle("/api/v1/budgets/", protected(budgetHandler(d.BudgetSvc, d.Log)))

	// Recurring transactions CRUD.
	mux.Handle("/api/v1/recurring", protected(recurringHandler(d.RecurringSvc, d.Log)))
	mux.Handle("/api/v1/recurring/", protected(recurringHandler(d.RecurringSvc, d.Log)))

	// Savings goals CRUD + deposit/withdraw.
	mux.Handle("/api/v1/goals", protected(goalsHandler(d.GoalSvc, d.Log)))
	mux.Handle("/api/v1/goals/", protected(goalsHandler(d.GoalSvc, d.Log)))

	// Data export.
	mux.Handle("/api/v1/export", protected(exportHandler(d.ExportSvc, d.Log)))

	return mux
}
