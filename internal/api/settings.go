package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type settingsResponse struct {
	BaseCurrency             string   `json:"base_currency"`
	DisplayCurrencies        []string `json:"display_currencies"`
	Language                 string   `json:"language"`
	IsAdmin                  bool     `json:"is_admin"`
	NotifyBudgetAlerts       bool     `json:"notify_budget_alerts"`
	NotifyRecurringReminders bool     `json:"notify_recurring_reminders"`
	NotifyWeeklySummary      bool     `json:"notify_weekly_summary"`
	NotifyGoalMilestones     bool     `json:"notify_goal_milestones"`
	StatsChartStyle          string   `json:"stats_chart_style"`
	AnimateNumbers           *bool    `json:"animate_numbers"`
	Theme                    string   `json:"theme"`
	HideAmounts              bool     `json:"hide_amounts"`
}

type notificationPrefsRequest struct {
	BudgetAlerts       *bool `json:"notify_budget_alerts"`
	RecurringReminders *bool `json:"notify_recurring_reminders"`
	WeeklySummary      *bool `json:"notify_weekly_summary"`
	GoalMilestones     *bool `json:"notify_goal_milestones"`
}

type uiPreferencesRequest struct {
	StatsChartStyle *string `json:"stats_chart_style"`
	AnimateNumbers  *bool   `json:"animate_numbers"`
	Theme           *string `json:"theme"`
	HideAmounts     *bool   `json:"hide_amounts"`
}

type patchSettingsRequest struct {
	DisplayCurrencies []string                  `json:"display_currencies"`
	Language          *string                   `json:"language"`
	NotificationPrefs *notificationPrefsRequest `json:"notification_preferences"`
	UIPreferences     *uiPreferencesRequest     `json:"ui_preferences"`
}

// settingsHandler handles GET and PATCH /api/v1/settings
func settingsHandler(userSvc *service.UserService, accountSvc *service.AccountService, adminUserID int64, devMode bool, log *slog.Logger) http.HandlerFunc {
	isAdmin := func(userID int64) bool {
		if devMode {
			return true
		}
		return adminUserID != 0 && userID == adminUserID
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		switch r.Method {
		case http.MethodGet:
			user, err := userSvc.GetByID(ctx, userID)
			if err != nil {
				writeError(w, log, err)
				return
			}
			// base_currency is derived from the default account.
			baseCurrency := "USD"
			if defaultAcc, accErr := accountSvc.GetDefault(ctx, userID); accErr == nil {
				baseCurrency = defaultAcc.CurrencyCode
			}
			writeJSON(w, http.StatusOK, settingsToResponse(user, baseCurrency, isAdmin(userID)))

		case http.MethodPatch:
			var req patchSettingsRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
				return
			}

			user, err := userSvc.GetByID(ctx, userID)
			if err != nil {
				writeError(w, log, err)
				return
			}

			if req.DisplayCurrencies != nil {
				user, err = userSvc.UpdateDisplayCurrencies(ctx, userID, req.DisplayCurrencies)
				if err != nil {
					writeError(w, log, err)
					return
				}
			}

			if req.Language != nil {
				user, err = userSvc.UpdateLanguage(ctx, userID, *req.Language)
				if err != nil {
					writeError(w, log, err)
					return
				}
			}

			if req.NotificationPrefs != nil {
				prefs := domain.NotificationPrefs{
					BudgetAlerts:       derefBool(req.NotificationPrefs.BudgetAlerts, user.NotifyBudgetAlerts),
					RecurringReminders: derefBool(req.NotificationPrefs.RecurringReminders, user.NotifyRecurringReminders),
					WeeklySummary:      derefBool(req.NotificationPrefs.WeeklySummary, user.NotifyWeeklySummary),
					GoalMilestones:     derefBool(req.NotificationPrefs.GoalMilestones, user.NotifyGoalMilestones),
				}
				user, err = userSvc.UpdateNotificationPreferences(ctx, userID, prefs)
				if err != nil {
					writeError(w, log, err)
					return
				}
			}

			if req.UIPreferences != nil {
				prefs := domain.UIPreferences{
					StatsChartStyle: user.StatsChartStyle,
					AnimateNumbers:  user.AnimateNumbers,
					Theme:           user.Theme,
					HideAmounts:     user.HideAmounts,
				}
				if req.UIPreferences.StatsChartStyle != nil {
					prefs.StatsChartStyle = *req.UIPreferences.StatsChartStyle
				}
				if req.UIPreferences.AnimateNumbers != nil {
					prefs.AnimateNumbers = req.UIPreferences.AnimateNumbers
				}
				if req.UIPreferences.Theme != nil {
					prefs.Theme = domain.ThemePref(*req.UIPreferences.Theme)
				}
				if req.UIPreferences.HideAmounts != nil {
					prefs.HideAmounts = *req.UIPreferences.HideAmounts
				}
				user, err = userSvc.UpdateUIPreferences(ctx, userID, prefs)
				if err != nil {
					writeError(w, log, err)
					return
				}
			}

			baseCurrency := "USD"
			if defaultAcc, accErr := accountSvc.GetDefault(ctx, userID); accErr == nil {
				baseCurrency = defaultAcc.CurrencyCode
			}
			writeJSON(w, http.StatusOK, settingsToResponse(user, baseCurrency, isAdmin(userID)))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func settingsToResponse(user *domain.User, baseCurrency string, isAdmin bool) settingsResponse {
	style := user.StatsChartStyle
	if style == "" {
		// Defensive default: should never happen since the column has a NOT NULL DEFAULT,
		// but treat empty string as the default style rather than leaking it to clients.
		style = domain.StatsChartStyleDonut
	}
	theme := string(user.Theme)
	if theme == "" {
		theme = string(domain.ThemeSystem)
	}
	return settingsResponse{
		BaseCurrency:             baseCurrency,
		DisplayCurrencies:        user.DisplayCurrencies,
		Language:                 string(user.Language),
		IsAdmin:                  isAdmin,
		NotifyBudgetAlerts:       user.NotifyBudgetAlerts,
		NotifyRecurringReminders: user.NotifyRecurringReminders,
		NotifyWeeklySummary:      user.NotifyWeeklySummary,
		NotifyGoalMilestones:     user.NotifyGoalMilestones,
		StatsChartStyle:          style,
		AnimateNumbers:           user.AnimateNumbers,
		Theme:                    theme,
		HideAmounts:              user.HideAmounts,
	}
}

// derefBool returns *ptr if ptr is non-nil, otherwise returns fallback.
func derefBool(ptr *bool, fallback bool) bool {
	if ptr != nil {
		return *ptr
	}
	return fallback
}
