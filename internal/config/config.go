package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	BotToken        string        `env:"BOT_TOKEN,required"`
	DatabaseURL     string        `env:"DATABASE_URL,required"`
	RedisURL        string        `env:"REDIS_URL,required"`
	LogLevel        string        `env:"LOG_LEVEL"          envDefault:"info"`
	MigrationsDir   string        `env:"MIGRATIONS_DIR"     envDefault:"db/migrations"`
	ExchangeRateTTL time.Duration `env:"EXCHANGE_RATE_TTL"  envDefault:"1h"`
	// API server configuration (used by cmd/api).
	APIPort        string `env:"API_PORT"        envDefault:"8080"`
	AllowedOrigins string `env:"ALLOWED_ORIGINS" envDefault:"*"`
	MiniAppURL     string `env:"MINI_APP_URL"    envDefault:""`
}

// Load parses environment variables into Config.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
