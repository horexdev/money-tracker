package domain

import "time"

// Language represents supported UI languages.
type Language string

const (
	LangEN Language = "en"
	LangRU Language = "ru"
	LangUK Language = "uk"
	LangBE Language = "be"
	LangKK Language = "kk"
	LangUZ Language = "uz"
	LangES Language = "es"
	LangDE Language = "de"
	LangIT Language = "it"
	LangFR Language = "fr"
	LangPT Language = "pt"
	LangNL Language = "nl"
	LangAR Language = "ar"
	LangTR Language = "tr"
	LangKO Language = "ko"
	LangMS Language = "ms"
	LangID Language = "id"
)

// supportedLanguages is the full set of accepted language codes.
var supportedLanguages = map[Language]struct{}{
	LangEN: {}, LangRU: {}, LangUK: {}, LangBE: {}, LangKK: {},
	LangUZ: {}, LangES: {}, LangDE: {}, LangIT: {}, LangFR: {},
	LangPT: {}, LangNL: {}, LangAR: {}, LangTR: {}, LangKO: {},
	LangMS: {}, LangID: {},
}

// ValidLanguage checks if the given string is a supported language code.
func ValidLanguage(code string) bool {
	_, ok := supportedLanguages[Language(code)]
	return ok
}

// NotificationPrefs holds a user's notification opt-in settings.
type NotificationPrefs struct {
	BudgetAlerts       bool
	RecurringReminders bool
	WeeklySummary      bool
	GoalMilestones     bool
}

// Stats chart styles selectable by the user on the Statistics screen.
const (
	StatsChartStyleDonut       = "donut"
	StatsChartStyleStackedBar  = "stacked_bar"
	StatsChartStyleDualBar     = "dual_bar"
	StatsChartStyleProfitBars  = "profit_bars"
)

// IsValidStatsChartStyle reports whether s is one of the supported chart styles.
func IsValidStatsChartStyle(s string) bool {
	switch s {
	case StatsChartStyleDonut, StatsChartStyleStackedBar, StatsChartStyleDualBar, StatsChartStyleProfitBars:
		return true
	}
	return false
}

// User represents a registered Telegram user.
type User struct {
	ID                       int64
	Username                 string
	FirstName                string
	LastName                 string
	Language                 Language
	DisplayCurrencies        []string
	CreatedAt                time.Time
	UpdatedAt                time.Time
	NotifyBudgetAlerts       bool
	NotifyRecurringReminders bool
	NotifyWeeklySummary      bool
	NotifyGoalMilestones     bool
	// UI preferences (stored on users table; see migration 00034).
	StatsChartStyle string
	// AnimateNumbers is nil when the user has not made an explicit choice;
	// the client falls back to the OS prefers-reduced-motion setting in that case.
	AnimateNumbers *bool
}

// DisplayName returns the best available name for the user.
func (u *User) DisplayName() string {
	if u.FirstName != "" {
		return u.FirstName
	}
	if u.Username != "" {
		return "@" + u.Username
	}
	return "User"
}
