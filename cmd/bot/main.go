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

	"github.com/horexdev/money-tracker/internal/config"
	"github.com/horexdev/money-tracker/internal/handler"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log := setupLogger(cfg.LogLevel)

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

	// Only UserService is needed for auto-registration.
	userRepo := repository.NewUserRepository(pool)
	userSvc := service.NewUserService(userRepo, log)

	b, err := bot.New(cfg.BotToken,
		bot.WithDefaultHandler(handler.DefaultHandler(cfg.MiniAppURL, log)),
		bot.WithMiddlewares(
			handler.LoggingMiddleware(log),
			handler.AutoRegisterMiddleware(userSvc, log),
		),
	)
	if err != nil {
		log.Error("failed to create bot", slog.String("error", err.Error()))
		os.Exit(1)
	}

	handler.RegisterAll(b, cfg.MiniAppURL, log)

	log.Info("bot started, waiting for updates")
	b.Start(ctx)
	log.Info("bot stopped")
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
