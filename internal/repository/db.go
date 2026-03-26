package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a *pgxpool.Pool from the given DSN.
// It validates the config and establishes an initial connection.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	return pgxpool.NewWithConfig(ctx, cfg)
}
