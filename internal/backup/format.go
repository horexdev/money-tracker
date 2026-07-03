package backup

import (
	"bytes"
	"encoding/json"
)

const (
	MoneyTrackerBackupFormat  = "MoneyTrackerBackup"
	MoneyTrackerBackupVersion = 1
)

type MoneyTrackerBackup struct {
	Format                string                       `json:"format"`
	Version               int                          `json:"version"`
	CreatedAtEpochMillis  int64                        `json:"created_at_epoch_millis"`
	Profiles              []BackupProfile              `json:"profiles"`
	ExchangeRateSnapshots []BackupExchangeRateSnapshot `json:"exchange_rate_snapshots"`
}

type BackupProfile struct {
	Ref                     string                        `json:"ref"`
	Label                   string                        `json:"label"`
	LanguageCode            string                        `json:"language_code"`
	DisplayCurrencyCodes    []string                      `json:"display_currency_codes"`
	NotificationPreferences BackupNotificationPreferences `json:"notification_preferences"`
	UIPreferences           BackupUIPreferences           `json:"ui_preferences"`
	CreatedAtEpochMillis    int64                         `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis    int64                         `json:"updated_at_epoch_millis"`
	Accounts                []BackupAccount               `json:"accounts"`
	Categories              []BackupCategory              `json:"categories"`
	Transactions            []BackupTransaction           `json:"transactions"`
	Transfers               []BackupTransfer              `json:"transfers"`
	Budgets                 []BackupBudget                `json:"budgets"`
	RecurringTransactions   []BackupRecurringTransaction  `json:"recurring_transactions"`
	SavingsGoals            []BackupSavingsGoal           `json:"savings_goals"`
	GoalTransactions        []BackupGoalTransaction       `json:"goal_transactions"`
	ExchangeRateOverrides   []BackupExchangeRateOverride  `json:"exchange_rate_overrides"`
	TransactionTemplates    []BackupTransactionTemplate   `json:"transaction_templates"`
}

type BackupNotificationPreferences struct {
	NotifyBudgetAlerts       bool `json:"notify_budget_alerts"`
	NotifyRecurringReminders bool `json:"notify_recurring_reminders"`
	NotifyWeeklySummary      bool `json:"notify_weekly_summary"`
	NotifyGoalMilestones     bool `json:"notify_goal_milestones"`
}

type BackupUIPreferences struct {
	StatsChartStyle string `json:"stats_chart_style"`
	AnimateNumbers  *bool  `json:"animate_numbers"`
	Theme           string `json:"theme"`
	HideAmounts     bool   `json:"hide_amounts"`
}

type BackupAccount struct {
	Ref                  string `json:"ref"`
	Name                 string `json:"name"`
	Icon                 string `json:"icon"`
	Color                string `json:"color"`
	Type                 string `json:"type"`
	CurrencyCode         string `json:"currency_code"`
	IsDefault            bool   `json:"is_default"`
	IncludeInTotal       bool   `json:"include_in_total"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis int64  `json:"updated_at_epoch_millis"`
}

type BackupCategory struct {
	Ref                  string `json:"ref"`
	Name                 string `json:"name"`
	Icon                 string `json:"icon"`
	Type                 string `json:"type"`
	Color                string `json:"color"`
	IsProtected          bool   `json:"is_protected"`
	UpdatedAtEpochMillis int64  `json:"updated_at_epoch_millis"`
	DeletedAtEpochMillis *int64 `json:"deleted_at_epoch_millis"`
}

type BackupTransaction struct {
	Ref                  string `json:"ref"`
	Type                 string `json:"type"`
	AmountCents          int64  `json:"amount_cents"`
	CategoryRef          string `json:"category_ref"`
	AccountRef           string `json:"account_ref"`
	Note                 string `json:"note"`
	CurrencyCode         string `json:"currency_code"`
	SnapshotDate         string `json:"snapshot_date"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
	IsAdjustment         bool   `json:"is_adjustment"`
}

type BackupTransfer struct {
	Ref                  string  `json:"ref"`
	FromAccountRef       string  `json:"from_account_ref"`
	ToAccountRef         string  `json:"to_account_ref"`
	AmountCents          int64   `json:"amount_cents"`
	FromCurrencyCode     string  `json:"from_currency_code"`
	ToCurrencyCode       string  `json:"to_currency_code"`
	ExchangeRateE8       int64   `json:"exchange_rate_e8"`
	Note                 string  `json:"note"`
	FromTransactionRef   *string `json:"from_transaction_ref"`
	ToTransactionRef     *string `json:"to_transaction_ref"`
	CreatedAtEpochMillis int64   `json:"created_at_epoch_millis"`
}

type BackupBudget struct {
	Ref                       string `json:"ref"`
	CategoryRef               string `json:"category_ref"`
	LimitCents                int64  `json:"limit_cents"`
	Period                    string `json:"period"`
	CurrencyCode              string `json:"currency_code"`
	NotifyAtPercent           int    `json:"notify_at_percent"`
	NotificationsEnabled      bool   `json:"notifications_enabled"`
	LastNotifiedPercent       int    `json:"last_notified_percent"`
	LastNotifiedAtEpochMillis *int64 `json:"last_notified_at_epoch_millis"`
	CreatedAtEpochMillis      int64  `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis      int64  `json:"updated_at_epoch_millis"`
}

type BackupRecurringTransaction struct {
	Ref                  string `json:"ref"`
	AccountRef           string `json:"account_ref"`
	CategoryRef          string `json:"category_ref"`
	Type                 string `json:"type"`
	AmountCents          int64  `json:"amount_cents"`
	CurrencyCode         string `json:"currency_code"`
	Note                 string `json:"note"`
	Frequency            string `json:"frequency"`
	NextRunAtEpochMillis int64  `json:"next_run_at_epoch_millis"`
	IsActive             bool   `json:"is_active"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis int64  `json:"updated_at_epoch_millis"`
}

type BackupSavingsGoal struct {
	Ref                  string  `json:"ref"`
	Name                 string  `json:"name"`
	TargetCents          int64   `json:"target_cents"`
	CurrentCents         int64   `json:"current_cents"`
	CurrencyCode         string  `json:"currency_code"`
	DeadlineDate         *string `json:"deadline_date"`
	AccountRef           *string `json:"account_ref"`
	CreatedAtEpochMillis int64   `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis int64   `json:"updated_at_epoch_millis"`
}

type BackupGoalTransaction struct {
	Ref                  string `json:"ref"`
	GoalRef              string `json:"goal_ref"`
	Type                 string `json:"type"`
	AmountCents          int64  `json:"amount_cents"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
}

type BackupExchangeRateSnapshot struct {
	SnapshotDate         string `json:"snapshot_date"`
	BaseCurrencyCode     string `json:"base_currency_code"`
	TargetCurrencyCode   string `json:"target_currency_code"`
	RateE8               int64  `json:"rate_e8"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
}

type BackupExchangeRateOverride struct {
	Ref                  string `json:"ref"`
	EffectiveDate        string `json:"effective_date"`
	BaseCurrencyCode     string `json:"base_currency_code"`
	TargetCurrencyCode   string `json:"target_currency_code"`
	RateE8               int64  `json:"rate_e8"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis int64  `json:"updated_at_epoch_millis"`
}

type BackupTransactionTemplate struct {
	Ref                  string `json:"ref"`
	Name                 string `json:"name"`
	Type                 string `json:"type"`
	AmountCents          int64  `json:"amount_cents"`
	AmountFixed          bool   `json:"amount_fixed"`
	CategoryRef          string `json:"category_ref"`
	AccountRef           string `json:"account_ref"`
	CurrencyCode         string `json:"currency_code"`
	Note                 string `json:"note"`
	SortOrder            int    `json:"sort_order"`
	CreatedAtEpochMillis int64  `json:"created_at_epoch_millis"`
	UpdatedAtEpochMillis int64  `json:"updated_at_epoch_millis"`
}

func NewEmptyBackup(createdAtEpochMillis int64) MoneyTrackerBackup {
	return MoneyTrackerBackup{
		Format:                MoneyTrackerBackupFormat,
		Version:               MoneyTrackerBackupVersion,
		CreatedAtEpochMillis:  createdAtEpochMillis,
		Profiles:              []BackupProfile{},
		ExchangeRateSnapshots: []BackupExchangeRateSnapshot{},
	}
}

func EncodeBackup(backup MoneyTrackerBackup) ([]byte, error) {
	if err := ValidateBackup(backup); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(backup); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
