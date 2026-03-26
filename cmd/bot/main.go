package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-telegram/bot"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"

	"github.com/horexdev/money-tracker/internal/config"
	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/handler"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/service"
)

func main() {
	// 1. Load configuration from environment variables.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 2. Setup structured logger.
	log := setupLogger(cfg.LogLevel)

	// 3. Run database migrations via goose.
	if err := runMigrations(cfg.DatabaseURL, cfg.MigrationsDir, log); err != nil {
		log.Error("migrations failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 4. Create context that cancels on OS signal.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 5. Connect to PostgreSQL via pgxpool.
	pool, err := repository.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed to connect to postgres", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 6. Connect to Redis.
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		// Fallback: treat RedisURL as bare host:port
		redisOpts = &redis.Options{Addr: cfg.RedisURL}
	}
	rdb := redis.NewClient(redisOpts)
	defer func() { _ = rdb.Close() }()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("failed to connect to redis", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 7. Build repositories.
	userRepo := repository.NewUserRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	catRepo := repository.NewCategoryRepository(pool)

	// 8. Build FSM store.
	fsmStore := fsm.NewStore(rdb)

	// 9. Build services.
	userSvc := service.NewUserService(userRepo, log)
	txSvc := service.NewTransactionService(txRepo, catRepo, log)
	statsSvc := service.NewStatsService(txRepo, log)
	exchangeSvc := service.NewExchangeService(service.NewRateAPIProvider(), rdb, cfg.ExchangeRateTTL, log)

	// 10. Build and configure the Telegram bot.
	b, err := bot.New(cfg.BotToken,
		bot.WithDefaultHandler(handler.DefaultHandler(fsmStore, txSvc, userSvc, log)),
		bot.WithMiddlewares(
			handler.LoggingMiddleware(log),
			handler.AutoRegisterMiddleware(userSvc, log),
		),
	)
	if err != nil {
		log.Error("failed to create bot", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 11. Register all command and callback handlers.
	handler.RegisterAll(b, fsmStore, userSvc, txSvc, statsSvc, exchangeSvc, log)

	log.Info("bot started, waiting for updates")

	// 12. Start the bot polling loop (blocks until ctx is cancelled).
	b.Start(ctx)

	log.Info("bot stopped")
}

// setupLogger creates a JSON logger for production or text logger for debug level.
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

// runMigrations applies all pending goose migrations.
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
