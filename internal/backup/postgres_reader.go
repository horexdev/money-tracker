package backup

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresReader struct {
	pool *pgxpool.Pool
}

func NewPostgresReader(pool *pgxpool.Pool) *PostgresReader {
	return &PostgresReader{pool: pool}
}

func (r *PostgresReader) ListUsers(ctx context.Context, userID *int64) ([]SourceUser, error) {
	var filter any
	if userID != nil {
		filter = *userID
	}
	rows, err := r.pool.Query(ctx, `
SELECT
    id,
    language,
    display_currencies,
    notify_budget_alerts,
    notify_recurring_reminders,
    notify_weekly_summary,
    notify_goal_milestones,
    stats_chart_style,
    animate_numbers,
    theme,
    hide_amounts,
    created_at,
    updated_at
FROM users
WHERE ($1::BIGINT IS NULL OR id = $1)
ORDER BY created_at ASC, id ASC
`, filter)
	if err != nil {
		return nil, fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []SourceUser
	for rows.Next() {
		var user SourceUser
		var animate pgtype.Bool
		if err := rows.Scan(
			&user.ID,
			&user.Language,
			&user.DisplayCurrencies,
			&user.NotifyBudgetAlerts,
			&user.NotifyRecurringReminders,
			&user.NotifyWeeklySummary,
			&user.NotifyGoalMilestones,
			&user.StatsChartStyle,
			&animate,
			&user.Theme,
			&user.HideAmounts,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.AnimateNumbers = optionalBool(animate)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (r *PostgresReader) LoadProfile(ctx context.Context, user SourceUser) (SourceProfileGraph, error) {
	graph := SourceProfileGraph{User: user}

	var err error
	if graph.Accounts, err = r.listAccounts(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.Categories, err = r.listCategories(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.Transactions, err = r.listTransactions(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.Transfers, err = r.listTransfers(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.Budgets, err = r.listBudgets(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.RecurringTransactions, err = r.listRecurringTransactions(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.SavingsGoals, err = r.listSavingsGoals(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.GoalTransactions, err = r.listGoalTransactions(ctx, user.ID); err != nil {
		return graph, err
	}
	if graph.TransactionTemplates, err = r.listTransactionTemplates(ctx, user.ID); err != nil {
		return graph, err
	}
	return graph, nil
}

func (r *PostgresReader) ListExchangeRateSnapshots(ctx context.Context) ([]SourceExchangeRateSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
    snapshot_date::TEXT,
    base_currency,
    target_currency,
    ROUND(rate * 100000000)::BIGINT AS rate_e8,
    created_at
FROM exchange_rate_snapshots
ORDER BY snapshot_date ASC, base_currency ASC, target_currency ASC
`)
	if err != nil {
		return nil, fmt.Errorf("query exchange rate snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []SourceExchangeRateSnapshot
	for rows.Next() {
		var snapshot SourceExchangeRateSnapshot
		if err := rows.Scan(
			&snapshot.SnapshotDate,
			&snapshot.BaseCurrencyCode,
			&snapshot.TargetCurrencyCode,
			&snapshot.RateE8,
			&snapshot.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exchange rate snapshot: %w", err)
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exchange rate snapshots: %w", err)
	}
	return snapshots, nil
}

func (r *PostgresReader) listAccounts(ctx context.Context, userID int64) ([]SourceAccount, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, icon, color, type::TEXT, currency_code, is_default, include_in_total, created_at, updated_at
FROM accounts
WHERE user_id = $1
ORDER BY is_default DESC, created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()

	var accounts []SourceAccount
	for rows.Next() {
		var account SourceAccount
		if err := rows.Scan(
			&account.ID,
			&account.Name,
			&account.Icon,
			&account.Color,
			&account.Type,
			&account.CurrencyCode,
			&account.IsDefault,
			&account.IncludeInTotal,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, account)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}
	return accounts, nil
}

func (r *PostgresReader) listCategories(ctx context.Context, userID int64) ([]SourceCategory, error) {
	rows, err := r.pool.Query(ctx, `
SELECT DISTINCT c.id, c.name, c.icon, c.type, c.color, c.is_protected, c.updated_at, c.deleted_at
FROM categories c
WHERE c.user_id = $1
   OR c.is_protected = TRUE
   OR EXISTS (SELECT 1 FROM transactions t WHERE t.user_id = $1 AND t.category_id = c.id)
   OR EXISTS (SELECT 1 FROM budgets b WHERE b.user_id = $1 AND b.category_id = c.id)
   OR EXISTS (SELECT 1 FROM recurring_transactions rt WHERE rt.user_id = $1 AND rt.category_id = c.id)
   OR EXISTS (SELECT 1 FROM transaction_templates tt WHERE tt.user_id = $1 AND tt.category_id = c.id)
ORDER BY c.id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}
	defer rows.Close()

	var categories []SourceCategory
	for rows.Next() {
		var category SourceCategory
		var deletedAt pgtype.Timestamptz
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Icon,
			&category.Type,
			&category.Color,
			&category.IsProtected,
			&category.UpdatedAt,
			&deletedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		category.DeletedAt = optionalTime(deletedAt)
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}
	return categories, nil
}

func (r *PostgresReader) listTransactions(ctx context.Context, userID int64) ([]SourceTransaction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, type::TEXT, amount_cents, category_id, account_id, note, currency_code, snapshot_date::TEXT, created_at, is_adjustment
FROM transactions
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []SourceTransaction
	for rows.Next() {
		var transaction SourceTransaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.Type,
			&transaction.AmountCents,
			&transaction.CategoryID,
			&transaction.AccountID,
			&transaction.Note,
			&transaction.CurrencyCode,
			&transaction.SnapshotDate,
			&transaction.CreatedAt,
			&transaction.IsAdjustment,
		); err != nil {
			return nil, fmt.Errorf("scan transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transactions: %w", err)
	}
	return transactions, nil
}

func (r *PostgresReader) listTransfers(ctx context.Context, userID int64) ([]SourceTransfer, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
    id,
    from_account_id,
    to_account_id,
    amount_cents,
    from_currency_code,
    to_currency_code,
    ROUND(exchange_rate * 100000000)::BIGINT AS exchange_rate_e8,
    note,
    from_tx_id,
    to_tx_id,
    created_at
FROM transfers
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query transfers: %w", err)
	}
	defer rows.Close()

	var transfers []SourceTransfer
	for rows.Next() {
		var transfer SourceTransfer
		var fromTxID pgtype.Int8
		var toTxID pgtype.Int8
		if err := rows.Scan(
			&transfer.ID,
			&transfer.FromAccountID,
			&transfer.ToAccountID,
			&transfer.AmountCents,
			&transfer.FromCurrencyCode,
			&transfer.ToCurrencyCode,
			&transfer.ExchangeRateE8,
			&transfer.Note,
			&fromTxID,
			&toTxID,
			&transfer.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		transfer.FromTxID = optionalInt64(fromTxID)
		transfer.ToTxID = optionalInt64(toTxID)
		transfers = append(transfers, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfers: %w", err)
	}
	return transfers, nil
}

func (r *PostgresReader) listBudgets(ctx context.Context, userID int64) ([]SourceBudget, error) {
	rows, err := r.pool.Query(ctx, `
SELECT
    id,
    category_id,
    limit_cents,
    period,
    currency_code,
    notify_at_percent,
    notifications_enabled,
    last_notified_percent,
    last_notified_at,
    created_at,
    updated_at
FROM budgets
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query budgets: %w", err)
	}
	defer rows.Close()

	var budgets []SourceBudget
	for rows.Next() {
		var budget SourceBudget
		var lastNotifiedAt pgtype.Timestamptz
		if err := rows.Scan(
			&budget.ID,
			&budget.CategoryID,
			&budget.LimitCents,
			&budget.Period,
			&budget.CurrencyCode,
			&budget.NotifyAtPercent,
			&budget.NotificationsEnabled,
			&budget.LastNotifiedPercent,
			&lastNotifiedAt,
			&budget.CreatedAt,
			&budget.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan budget: %w", err)
		}
		budget.LastNotifiedAt = optionalTime(lastNotifiedAt)
		budgets = append(budgets, budget)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate budgets: %w", err)
	}
	return budgets, nil
}

func (r *PostgresReader) listRecurringTransactions(ctx context.Context, userID int64) ([]SourceRecurringTransaction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, account_id, category_id, type::TEXT, amount_cents, currency_code, note, frequency, next_run_at, is_active, created_at, updated_at
FROM recurring_transactions
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query recurring transactions: %w", err)
	}
	defer rows.Close()

	var recurring []SourceRecurringTransaction
	for rows.Next() {
		var row SourceRecurringTransaction
		if err := rows.Scan(
			&row.ID,
			&row.AccountID,
			&row.CategoryID,
			&row.Type,
			&row.AmountCents,
			&row.CurrencyCode,
			&row.Note,
			&row.Frequency,
			&row.NextRunAt,
			&row.IsActive,
			&row.CreatedAt,
			&row.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recurring transaction: %w", err)
		}
		recurring = append(recurring, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recurring transactions: %w", err)
	}
	return recurring, nil
}

func (r *PostgresReader) listSavingsGoals(ctx context.Context, userID int64) ([]SourceSavingsGoal, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, target_cents, current_cents, currency_code, deadline::TEXT, account_id, created_at, updated_at
FROM savings_goals
WHERE user_id = $1
ORDER BY created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query savings goals: %w", err)
	}
	defer rows.Close()

	var goals []SourceSavingsGoal
	for rows.Next() {
		var goal SourceSavingsGoal
		var deadline pgtype.Text
		var accountID pgtype.Int8
		if err := rows.Scan(
			&goal.ID,
			&goal.Name,
			&goal.TargetCents,
			&goal.CurrentCents,
			&goal.CurrencyCode,
			&deadline,
			&accountID,
			&goal.CreatedAt,
			&goal.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan savings goal: %w", err)
		}
		goal.DeadlineDate = optionalText(deadline)
		goal.AccountID = optionalInt64(accountID)
		goals = append(goals, goal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate savings goals: %w", err)
	}
	return goals, nil
}

func (r *PostgresReader) listGoalTransactions(ctx context.Context, userID int64) ([]SourceGoalTransaction, error) {
	rows, err := r.pool.Query(ctx, `
SELECT gt.id, gt.goal_id, gt.type, gt.amount_cents, gt.created_at
FROM goal_transactions gt
JOIN savings_goals g ON g.id = gt.goal_id AND g.user_id = gt.user_id
WHERE gt.user_id = $1
ORDER BY gt.created_at ASC, gt.id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query goal transactions: %w", err)
	}
	defer rows.Close()

	var transactions []SourceGoalTransaction
	for rows.Next() {
		var transaction SourceGoalTransaction
		if err := rows.Scan(
			&transaction.ID,
			&transaction.GoalID,
			&transaction.Type,
			&transaction.AmountCents,
			&transaction.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan goal transaction: %w", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate goal transactions: %w", err)
	}
	return transactions, nil
}

func (r *PostgresReader) listTransactionTemplates(ctx context.Context, userID int64) ([]SourceTransactionTemplate, error) {
	rows, err := r.pool.Query(ctx, `
SELECT id, name, type::TEXT, amount_cents, amount_fixed, category_id, account_id, currency_code, note, sort_order, created_at, updated_at
FROM transaction_templates
WHERE user_id = $1
ORDER BY sort_order ASC, created_at ASC, id ASC
`, userID)
	if err != nil {
		return nil, fmt.Errorf("query transaction templates: %w", err)
	}
	defer rows.Close()

	var templates []SourceTransactionTemplate
	for rows.Next() {
		var template SourceTransactionTemplate
		if err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Type,
			&template.AmountCents,
			&template.AmountFixed,
			&template.CategoryID,
			&template.AccountID,
			&template.CurrencyCode,
			&template.Note,
			&template.SortOrder,
			&template.CreatedAt,
			&template.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transaction template: %w", err)
		}
		templates = append(templates, template)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transaction templates: %w", err)
	}
	return templates, nil
}

func optionalTime(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func optionalInt64(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}

func optionalBool(value pgtype.Bool) *bool {
	if !value.Valid {
		return nil
	}
	v := value.Bool
	return &v
}

func optionalText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}
