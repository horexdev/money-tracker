package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/config"
)

// setEnv sets multiple env vars for the test and restores them on cleanup.
func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

// Aliases that allow the test to unset env vars (which testing.T.Setenv does
// not support) while still using a single import line.
var (
	osLookupEnv = os.LookupEnv
	osUnsetenv  = os.Unsetenv
	osSetenv    = os.Setenv
)

func TestLoad_RequiredFields(t *testing.T) {
	setEnv(t, map[string]string{
		"BOT_TOKEN":    "tok",
		"DATABASE_URL": "postgres://x",
		"REDIS_URL":    "redis://y",
	})
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "tok", cfg.BotToken)
	assert.Equal(t, "postgres://x", cfg.DatabaseURL)
	assert.Equal(t, "redis://y", cfg.RedisURL)
}

func TestLoad_AppliesDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"BOT_TOKEN":    "tok",
		"DATABASE_URL": "postgres://x",
		"REDIS_URL":    "redis://y",
	})
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "db/migrations", cfg.MigrationsDir)
	assert.Equal(t, time.Hour, cfg.ExchangeRateTTL)
	assert.Equal(t, "8080", cfg.APIPort)
	assert.Equal(t, "*", cfg.AllowedOrigins)
	assert.Empty(t, cfg.MiniAppURL)
	assert.Equal(t, int64(0), cfg.AdminUserID)
	assert.False(t, cfg.DevMode)
	assert.Equal(t, "en", cfg.DevLang)
}

func TestLoad_OverridesDefaults(t *testing.T) {
	setEnv(t, map[string]string{
		"BOT_TOKEN":         "tok",
		"DATABASE_URL":      "postgres://x",
		"REDIS_URL":         "redis://y",
		"LOG_LEVEL":         "debug",
		"MIGRATIONS_DIR":    "/tmp/migrations",
		"EXCHANGE_RATE_TTL": "30m",
		"API_PORT":          "9090",
		"ALLOWED_ORIGINS":   "https://app.example.com",
		"MINI_APP_URL":      "https://t.me/app",
		"ADMIN_USER_ID":     "12345",
		"DEV_MODE":          "true",
		"DEV_LANG":          "ru",
	})
	cfg, err := config.Load()
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "/tmp/migrations", cfg.MigrationsDir)
	assert.Equal(t, 30*time.Minute, cfg.ExchangeRateTTL)
	assert.Equal(t, "9090", cfg.APIPort)
	assert.Equal(t, "https://app.example.com", cfg.AllowedOrigins)
	assert.Equal(t, "https://t.me/app", cfg.MiniAppURL)
	assert.Equal(t, int64(12345), cfg.AdminUserID)
	assert.True(t, cfg.DevMode)
	assert.Equal(t, "ru", cfg.DevLang)
}

func TestLoad_ErrorsOnMissingRequired(t *testing.T) {
	saved := map[string]string{}
	for _, k := range []string{"BOT_TOKEN", "DATABASE_URL", "REDIS_URL"} {
		if v, ok := osLookupEnv(k); ok {
			saved[k] = v
		}
		osUnsetenv(k)
	}
	t.Cleanup(func() {
		for k, v := range saved {
			osSetenv(k, v)
		}
	})

	_, err := config.Load()
	assert.Error(t, err)
}
