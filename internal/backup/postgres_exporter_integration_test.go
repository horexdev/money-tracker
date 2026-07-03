//go:build integration

package backup_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/backup"
	"github.com/horexdev/money-tracker/internal/testutil"
)

func TestPostgresExporterWritesBackupFile(t *testing.T) {
	pool := testutil.OpenTestPool(t)
	testutil.CleanupTables(t, pool)
	t.Cleanup(func() { testutil.CleanupTables(t, pool) })

	ctx := context.Background()
	userID := int64(4242424242)
	createdAt := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	_, err := pool.Exec(ctx, `
INSERT INTO users (
    id, username, first_name, last_name, language, display_currencies,
    notify_budget_alerts, notify_recurring_reminders, notify_weekly_summary, notify_goal_milestones,
    stats_chart_style, animate_numbers, theme, hide_amounts, created_at, updated_at
) VALUES (
    $1, 'tg_user', 'Telegram', 'Person', 'ru', 'USD,EUR',
    TRUE, TRUE, FALSE, TRUE,
    'donut', TRUE, 'dark', TRUE, $2, $2
)`, userID, createdAt)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var mainAccountID, savingsAccountID int64
	err = pool.QueryRow(ctx, `
INSERT INTO accounts (user_id, name, icon, color, type, currency_code, is_default, include_in_total, created_at, updated_at)
VALUES ($1, 'Main', 'wallet', '#6366f1', 'checking', 'USD', TRUE, TRUE, $2, $2)
RETURNING id`, userID, createdAt).Scan(&mainAccountID)
	if err != nil {
		t.Fatalf("insert main account: %v", err)
	}
	err = pool.QueryRow(ctx, `
INSERT INTO accounts (user_id, name, icon, color, type, currency_code, is_default, include_in_total, created_at, updated_at)
VALUES ($1, 'Savings', 'piggy-bank', '#22c55e', 'savings', 'EUR', FALSE, TRUE, $2, $2)
RETURNING id`, userID, createdAt).Scan(&savingsAccountID)
	if err != nil {
		t.Fatalf("insert savings account: %v", err)
	}

	var foodCategoryID, transferCategoryID int64
	err = pool.QueryRow(ctx, `
INSERT INTO categories (user_id, name, icon, type, color, updated_at)
VALUES ($1, 'Food', 'fork-knife', 'expense', '#ef4444', $2)
RETURNING id`, userID, createdAt).Scan(&foodCategoryID)
	if err != nil {
		t.Fatalf("insert category: %v", err)
	}
	err = pool.QueryRow(ctx, `SELECT id FROM categories WHERE user_id IS NULL AND type = 'transfer' LIMIT 1`).Scan(&transferCategoryID)
	if err != nil {
		t.Fatalf("get transfer category: %v", err)
	}

	var expenseTxID, transferOutTxID, transferInTxID int64
	err = pool.QueryRow(ctx, `
INSERT INTO transactions (user_id, type, amount_cents, category_id, account_id, note, currency_code, snapshot_date, created_at)
VALUES ($1, 'expense', 1250, $2, $3, 'Lunch', 'USD', '2026-07-01', $4)
RETURNING id`, userID, foodCategoryID, mainAccountID, createdAt).Scan(&expenseTxID)
	if err != nil {
		t.Fatalf("insert expense transaction: %v", err)
	}
	err = pool.QueryRow(ctx, `
INSERT INTO transactions (user_id, type, amount_cents, category_id, account_id, note, currency_code, snapshot_date, created_at)
VALUES ($1, 'expense', 5000, $2, $3, 'Transfer out', 'USD', '2026-07-01', $4)
RETURNING id`, userID, transferCategoryID, mainAccountID, createdAt.Add(time.Minute)).Scan(&transferOutTxID)
	if err != nil {
		t.Fatalf("insert transfer out transaction: %v", err)
	}
	err = pool.QueryRow(ctx, `
INSERT INTO transactions (user_id, type, amount_cents, category_id, account_id, note, currency_code, snapshot_date, created_at)
VALUES ($1, 'income', 4600, $2, $3, 'Transfer in', 'EUR', '2026-07-01', $4)
RETURNING id`, userID, transferCategoryID, savingsAccountID, createdAt.Add(2*time.Minute)).Scan(&transferInTxID)
	if err != nil {
		t.Fatalf("insert transfer in transaction: %v", err)
	}
	_ = expenseTxID

	_, err = pool.Exec(ctx, `
INSERT INTO transfers (
    user_id, from_account_id, to_account_id, amount_cents,
    from_currency_code, to_currency_code, exchange_rate, note, created_at, from_tx_id, to_tx_id
) VALUES ($1, $2, $3, 5000, 'USD', 'EUR', 0.92000000, 'Transfer', $4, $5, $6)`,
		userID, mainAccountID, savingsAccountID, createdAt.Add(3*time.Minute), transferOutTxID, transferInTxID)
	if err != nil {
		t.Fatalf("insert transfer: %v", err)
	}

	_, err = pool.Exec(ctx, `
INSERT INTO budgets (user_id, category_id, limit_cents, period, currency_code, notify_at_percent, notifications_enabled, last_notified_percent, created_at, updated_at)
VALUES ($1, $2, 50000, 'monthly', 'USD', 80, TRUE, 0, $3, $3)`, userID, foodCategoryID, createdAt)
	if err != nil {
		t.Fatalf("insert budget: %v", err)
	}
	_, err = pool.Exec(ctx, `
INSERT INTO recurring_transactions (user_id, account_id, category_id, type, amount_cents, currency_code, note, frequency, next_run_at, created_at, updated_at)
VALUES ($1, $2, $3, 'expense', 90000, 'USD', 'Rent', 'monthly', $4, $4, $4)`,
		userID, mainAccountID, foodCategoryID, createdAt)
	if err != nil {
		t.Fatalf("insert recurring: %v", err)
	}
	var goalID int64
	err = pool.QueryRow(ctx, `
INSERT INTO savings_goals (user_id, name, target_cents, current_cents, currency_code, deadline, account_id, created_at, updated_at)
VALUES ($1, 'Trip', 200000, 10000, 'EUR', '2026-12-31', $2, $3, $3)
RETURNING id`, userID, savingsAccountID, createdAt).Scan(&goalID)
	if err != nil {
		t.Fatalf("insert savings goal: %v", err)
	}
	_, err = pool.Exec(ctx, `
INSERT INTO goal_transactions (goal_id, user_id, type, amount_cents, created_at)
VALUES ($1, $2, 'deposit', 10000, $3)`, goalID, userID, createdAt)
	if err != nil {
		t.Fatalf("insert goal transaction: %v", err)
	}
	_, err = pool.Exec(ctx, `
INSERT INTO transaction_templates (user_id, name, type, amount_cents, amount_fixed, category_id, account_id, currency_code, note, sort_order, created_at, updated_at)
VALUES ($1, 'Lunch', 'expense', 1250, TRUE, $2, $3, 'USD', 'Template', 0, $4, $4)`,
		userID, foodCategoryID, mainAccountID, createdAt)
	if err != nil {
		t.Fatalf("insert template: %v", err)
	}
	_, err = pool.Exec(ctx, `
INSERT INTO exchange_rate_snapshots (snapshot_date, base_currency, target_currency, rate, created_at)
VALUES ('2026-07-01', 'USD', 'EUR', 0.92000000, $1)`, createdAt)
	if err != nil {
		t.Fatalf("insert exchange snapshot: %v", err)
	}

	exporter := backup.NewExporter(backup.NewPostgresReader(pool))
	result, err := exporter.Export(ctx, &userID)
	if err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	data, err := backup.EncodeBackup(result)
	if err != nil {
		t.Fatalf("EncodeBackup() error = %v", err)
	}

	path := filepath.Join(t.TempDir(), "moneytracker.mtbackup.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	encoded := string(written)
	for _, forbidden := range []string{"4242424242", "tg_user", "Telegram", "Person", `"id"`, `"user_id"`, `"account_id"`, `"category_id"`} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("backup file contains forbidden value/key %q:\n%s", forbidden, encoded)
		}
	}
	if !strings.Contains(encoded, `"format": "MoneyTrackerBackup"`) {
		t.Fatalf("backup file does not contain MoneyTrackerBackup format:\n%s", encoded)
	}
}
