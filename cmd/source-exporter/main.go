package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/horexdev/money-tracker/internal/backup"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := parseFlags(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	reader := backup.NewPostgresReader(pool)
	exporter := backup.NewExporter(reader)
	out, err := exporter.Export(ctx, cfg.userID.ptr())
	if err != nil {
		return fmt.Errorf("export backup: %w", err)
	}

	data, err := backup.EncodeBackup(out)
	if err != nil {
		return fmt.Errorf("encode backup: %w", err)
	}
	if err := writeOutput(cfg.output, data); err != nil {
		return err
	}
	return nil
}

type config struct {
	databaseURL string
	output      string
	userID      optionalUserID
}

func parseFlags(args []string) (config, error) {
	var cfg config
	fs := flag.NewFlagSet("source-exporter", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&cfg.databaseURL, "database-url", os.Getenv("DATABASE_URL"), "PostgreSQL DSN; defaults to DATABASE_URL")
	fs.StringVar(&cfg.output, "output", "", "output .mtbackup.json path, or - for stdout")
	fs.Var(&cfg.userID, "user-id", "optional operator filter by source users.id")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if cfg.databaseURL == "" {
		return cfg, fmt.Errorf("--database-url or DATABASE_URL is required")
	}
	if cfg.output == "" {
		return cfg, fmt.Errorf("--output is required")
	}
	return cfg, nil
}

type optionalUserID struct {
	value int64
	set   bool
}

func (u *optionalUserID) String() string {
	if !u.set {
		return ""
	}
	return strconv.FormatInt(u.value, 10)
}

func (u *optionalUserID) Set(value string) error {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user id: %w", err)
	}
	if parsed <= 0 {
		return fmt.Errorf("user id must be positive")
	}
	u.value = parsed
	u.set = true
	return nil
}

func (u optionalUserID) ptr() *int64 {
	if !u.set {
		return nil
	}
	v := u.value
	return &v
}

func writeOutput(path string, data []byte) error {
	if path == "-" {
		_, err := os.Stdout.Write(data)
		return err
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
