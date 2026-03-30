package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

// defaultAccountCurrency maps Telegram language_code to the most common currency
// for that locale. Used when creating the first default account for a new user,
// before the frontend has had a chance to persist a preference.
var defaultAccountCurrencies = map[string]string{
	"en": "USD", "ru": "RUB", "uk": "UAH", "be": "BYN",
	"kk": "KZT", "uz": "UZS", "tr": "TRY", "ar": "SAR",
	"es": "EUR", "pt": "BRL", "fr": "EUR", "de": "EUR",
	"it": "EUR", "nl": "EUR", "ko": "KRW", "ms": "MYR", "id": "IDR",
}

func localizedAccountCurrency(lang string) string {
	if c, ok := defaultAccountCurrencies[lang]; ok {
		return c
	}
	return "USD"
}

// defaultAccountName returns a localized name for the first default account.
// Falls back to English if the language code is not recognized.
var defaultAccountNames = map[string]string{
	"en": "Main account",
	"ru": "Основной счёт",
	"uk": "Основний рахунок",
	"be": "Асноўны рахунак",
	"kk": "Негізгі шот",
	"uz": "Asosiy hisob",
	"es": "Cuenta principal",
	"de": "Hauptkonto",
	"it": "Conto principale",
	"fr": "Compte principal",
	"pt": "Conta principal",
	"nl": "Hoofdrekening",
	"ar": "الحساب الرئيسي",
	"tr": "Ana hesap",
	"ko": "주 계좌",
	"ms": "Akaun utama",
	"id": "Rekening utama",
}

func localizedAccountName(lang string) string {
	if name, ok := defaultAccountNames[lang]; ok {
		return name
	}
	return defaultAccountNames["en"]
}

// userEnsurer wraps UserService to satisfy the ensurer interface required by authMiddleware.
// It also creates a default account for first-time users.
type userEnsurer struct {
	svc        *service.UserService
	accountSvc *service.AccountService
	log        *slog.Logger
}

func (u *userEnsurer) ensureUser(ctx context.Context, tgUser TelegramUser) error {
	u.log.InfoContext(ctx, "ensureUser: tg profile",
		slog.Int64("user_id", tgUser.ID),
		slog.String("language_code", tgUser.LanguageCode),
		slog.String("first_name", tgUser.FirstName),
	)
	user, err := u.svc.Upsert(ctx, &domain.User{
		ID:           tgUser.ID,
		Username:     tgUser.Username,
		FirstName:    tgUser.FirstName,
		LastName:     tgUser.LastName,
		Language:     domain.Language(tgUser.LanguageCode),
		CurrencyCode: localizedAccountCurrency(tgUser.LanguageCode),
	})
	if err != nil {
		return err
	}

	accounts, err := u.accountSvc.List(ctx, user.ID)
	if err != nil {
		u.log.WarnContext(ctx, "ensureUser: failed to list accounts",
			slog.Int64("user_id", user.ID),
			slog.String("error", err.Error()),
		)
		return nil
	}
	if len(accounts) > 0 {
		return nil
	}

	// Prefer the persisted language; fall back to the Telegram-reported one.
	lang := string(user.Language)
	if lang == "" {
		lang = tgUser.LanguageCode
	}
	name := localizedAccountName(lang)
	// Derive currency from the Telegram language_code so the first account
	// gets a sensible default even before the frontend persists a preference.
	currency := localizedAccountCurrency(lang)

	if _, err := u.accountSvc.Create(ctx, user.ID, name, "wallet", "#6366f1", domain.AccountTypeChecking, currency, true); err != nil {
		u.log.WarnContext(ctx, "ensureUser: failed to create default account",
			slog.Int64("user_id", user.ID),
			slog.String("error", err.Error()),
		)
	}
	return nil
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
	AccountSvc     *service.AccountService
	TransferSvc    *service.TransferService
	AdminSvc       *service.AdminService
	BotToken       string
	AllowedOrigins string
	AdminUserID    int64
	// DevMode enables the Telegram auth bypass for local development.
	// When true, X-Telegram-Init-Data: dev:<user_id> is accepted without HMAC validation.
	DevMode bool
	DevLang string
	Log     *slog.Logger
}

// NewServer builds and returns the HTTP handler for the API server.
func NewServer(d Deps) http.Handler {
	mux := http.NewServeMux()

	eu := &userEnsurer{svc: d.UserSvc, accountSvc: d.AccountSvc, log: d.Log}

	auth := authMiddleware(d.BotToken, d.DevMode, d.DevLang, eu, d.Log)
	cors := corsMiddleware(d.AllowedOrigins)
	logging := loggingMiddleware(d.Log)

	// protected wraps a handler with logging + CORS + auth.
	protected := func(h http.HandlerFunc) http.Handler {
		return chain(h, logging, cors, auth)
	}

	// adminProtected wraps a handler with logging + CORS + auth + admin check.
	admin := adminMiddleware(d.AdminUserID, d.DevMode)
	adminProtected := func(h http.HandlerFunc) http.Handler {
		return chain(h, logging, cors, auth, admin)
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
	mux.Handle("/api/v1/transactions", protected(transactionHandler(d.TxSvc, d.UserSvc, d.ExchangeSvc, d.Log)))
	mux.Handle("/api/v1/transactions/", protected(transactionHandler(d.TxSvc, d.UserSvc, d.ExchangeSvc, d.Log)))
	mux.Handle("/api/v1/stats", protected(statsHandler(d.StatsSvc, d.Log)))
	mux.Handle("/api/v1/settings", protected(settingsHandler(d.UserSvc, d.AdminUserID, d.DevMode, d.Log)))

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

	// Accounts CRUD.
	mux.Handle("/api/v1/accounts", protected(accountsHandler(d.AccountSvc, d.Log)))
	mux.Handle("/api/v1/accounts/", protected(accountsHandler(d.AccountSvc, d.Log)))

	// Transfers.
	mux.Handle("/api/v1/transfers", protected(transfersHandler(d.TransferSvc, d.Log)))
	mux.Handle("/api/v1/transfers/", protected(transfersHandler(d.TransferSvc, d.Log)))

	// Exchange rate lookup.
	mux.Handle("/api/v1/exchange/rate", protected(exchangeRateHandler(d.ExchangeSvc, d.Log)))

	// Data export.
	mux.Handle("/api/v1/export", protected(exportHandler(d.ExportSvc, d.Log)))

	// User data reset.
	mux.Handle("/api/v1/user/data", protected(userDataHandler(d.UserSvc, d.Log)))

	// Admin endpoints.
	mux.Handle("/api/v1/admin/users", adminProtected(adminUsersHandler(d.AdminSvc, d.Log)))
	mux.Handle("/api/v1/admin/stats", adminProtected(adminStatsHandler(d.AdminSvc, d.Log)))
	mux.Handle("/api/v1/admin/users/data", adminProtected(adminResetAllHandler(d.AdminSvc, d.UserSvc, d.AccountSvc, d.Log)))
	mux.Handle("/api/v1/admin/users/", adminProtected(adminResetUserHandler(d.AdminSvc, d.UserSvc, d.AccountSvc, d.Log)))

	return mux
}
