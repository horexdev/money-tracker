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
}

type notificationPrefsRequest struct {
	BudgetAlerts       *bool `json:"notify_budget_alerts"`
	RecurringReminders *bool `json:"notify_recurring_reminders"`
	WeeklySummary      *bool `json:"notify_weekly_summary"`
	GoalMilestones     *bool `json:"notify_goal_milestones"`
}

type patchSettingsRequest struct {
	BaseCurrency         *string                   `json:"base_currency"`
	DisplayCurrencies    []string                  `json:"display_currencies"`
	Language             *string                   `json:"language"`
	NotificationPrefs    *notificationPrefsRequest `json:"notification_preferences"`
}

// settingsHandler handles GET and PATCH /api/v1/settings
func settingsHandler(userSvc *service.UserService, adminUserID int64, devMode bool, log *slog.Logger) http.HandlerFunc {
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
			writeJSON(w, http.StatusOK, userToSettingsResponse(user, isAdmin(userID)))

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

			if req.BaseCurrency != nil {
				user, err = userSvc.UpdateCurrency(ctx, userID, *req.BaseCurrency)
				if err != nil {
					writeError(w, log, err)
					return
				}
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

			writeJSON(w, http.StatusOK, userToSettingsResponse(user, isAdmin(userID)))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func userToSettingsResponse(user *domain.User, isAdmin bool) settingsResponse {
	return settingsResponse{
		BaseCurrency:             user.CurrencyCode,
		DisplayCurrencies:        user.DisplayCurrencies,
		Language:                 string(user.Language),
		IsAdmin:                  isAdmin,
		NotifyBudgetAlerts:       user.NotifyBudgetAlerts,
		NotifyRecurringReminders: user.NotifyRecurringReminders,
		NotifyWeeklySummary:      user.NotifyWeeklySummary,
		NotifyGoalMilestones:     user.NotifyGoalMilestones,
	}
}

// derefBool returns *ptr if ptr is non-nil, otherwise returns fallback.
func derefBool(ptr *bool, fallback bool) bool {
	if ptr != nil {
		return *ptr
	}
	return fallback
}
