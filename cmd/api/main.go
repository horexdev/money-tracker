package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/config"
	"github.com/horexdev/money-tracker/internal/notify"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/scheduler"
	"github.com/horexdev/money-tracker/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := setupLogger(cfg.LogLevel)

	if cfg.DevMode {
		log.Warn("DEV_MODE is enabled — Telegram auth bypass is active. DO NOT use in production.")
	}

	if err := runMigrations(cfg.DatabaseURL, cfg.MigrationsDir, log); err != nil {
		log.Error("migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		redisOpts = &redis.Options{Addr: cfg.RedisURL}
	}
	rdb := redis.NewClient(redisOpts)
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Repositories.
	userRepo := repository.NewUserRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	catRepo := repository.NewCategoryRepository(pool)
	budgetRepo := repository.NewBudgetRepository(pool)
	recurringRepo := repository.NewRecurringRepository(pool)
	goalRepo := repository.NewSavingsGoalRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	transferRepo := repository.NewTransferRepository(pool)
	adminRepo := repository.NewAdminRepository(pool)
	snapshotRepo := repository.NewExchangeSnapshotRepository(pool)

	// Services.
	userSvc := service.NewUserService(userRepo, log)
	txSvc := service.NewTransactionService(txRepo, catRepo, log)
	statsSvc := service.NewStatsService(txRepo, log)
	exchangeSvc := service.NewExchangeService(service.NewRateAPIProvider(), rdb, cfg.ExchangeRateTTL, log)
	categorySvc := service.NewCategoryService(catRepo, log)
	budgetSvc := service.NewBudgetService(budgetRepo, txRepo, userRepo, log)
	recurringSvc := service.NewRecurringService(recurringRepo, txRepo, accountRepo, log)
	goalSvc := service.NewSavingsGoalService(goalRepo, txRepo, catRepo, accountRepo, log)
	exportSvc := service.NewExportService(txRepo, log)
	accountSvc := service.NewAccountService(accountRepo, goalRepo, log)
	transferSvc := service.NewTransferService(transferRepo, accountRepo, goalRepo, txRepo, catRepo, log)
	adjustSvc := service.NewAdjustmentService(txRepo, accountRepo, catRepo, log)
	adminSvc := service.NewAdminService(adminRepo, log)
	snapshotSvc := service.NewSnapshotService(snapshotRepo, service.NewRateAPIProvider(), log)

	// Wire budget notifier if a bot token is configured.
	if cfg.BotToken != "" {
		notifier := notify.NewTelegramNotifier(cfg.BotToken, log)
		budgetSvc.WithNotifier(notifier)
	}

	// Background scheduler for recurring transactions and budget alerts.
	sched := scheduler.New(recurringSvc, budgetSvc, snapshotSvc, log, 1*time.Minute)
	go sched.Run(ctx)

	handler := api.NewServer(api.Deps{
		UserSvc:        userSvc,
		TxSvc:          txSvc,
		StatsSvc:       statsSvc,
		ExchangeSvc:    exchangeSvc,
		CategorySvc:    categorySvc,
		BudgetSvc:      budgetSvc,
		RecurringSvc:   recurringSvc,
		GoalSvc:        goalSvc,
		ExportSvc:      exportSvc,
		AccountSvc:     accountSvc,
		TransferSvc:    transferSvc,
		AdjustSvc:      adjustSvc,
		AdminSvc:       adminSvc,
		SnapshotSvc:    snapshotSvc,
		BotToken:       cfg.BotToken,
		AllowedOrigins: cfg.AllowedOrigins,
		AdminUserID:    cfg.AdminUserID,
		DevMode:        cfg.DevMode,
		DevLang:        cfg.DevLang,
		Log:            log,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      handler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("api server started", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("api server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down api server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", slog.String("error", err.Error()))
	}
	log.Info("api server stopped")
}

func setupLogger(level string) *slog.Logger {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lvl}
	if lvl == slog.LevelDebug {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func runMigrations(dsn, dir string, log *slog.Logger) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	log.Info("running migrations", slog.String("dir", dir))
	if err := goose.Up(db, dir); err != nil {
		return err
	}
	log.Info("migrations applied successfully")
	return nil
}
