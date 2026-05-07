package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildSettingsHandler(userRepo *mocks.MockUserStorer, accountRepo *mocks.MockAccountStorer) http.HandlerFunc {
	log := testutil.TestLogger()
	userSvc := service.NewUserService(userRepo, log)
	accountSvc := service.NewAccountService(accountRepo, nil, log)
	return api.SettingsHandlerForTest(userSvc, accountSvc, 0, log)
}

func TestSettingsHandler_GET(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:                1,
		Language:          domain.LangEN,
		DisplayCurrencies: []string{"EUR"},
	}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "USD", resp["base_currency"])
	assert.Equal(t, "en", resp["language"])
}

func TestSettingsHandler_GET_BaseCurrencyFromDefaultAccount(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:       1,
		Language: domain.LangEN,
	}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	// Default account has EUR — this should be returned as base_currency
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 2, CurrencyCode: "EUR"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "EUR", resp["base_currency"])
}

func TestSettingsHandler_PATCH_TooManyDisplayCurrencies(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1}, nil)

	accountRepo := &mocks.MockAccountStorer{}

	h := buildSettingsHandler(userRepo, accountRepo)
	body := `{"display_currencies":["USD","EUR","GBP","JPY"]}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSettingsHandler_UnsupportedMethod(t *testing.T) {
	h := buildSettingsHandler(&mocks.MockUserStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestSettingsHandler_GET_IncludesNotificationPrefs(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:                       1,
		Language:                 domain.LangEN,
		NotifyBudgetAlerts:       true,
		NotifyRecurringReminders: false,
		NotifyWeeklySummary:      true,
		NotifyGoalMilestones:     false,
	}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, true, resp["notify_budget_alerts"])
	assert.Equal(t, false, resp["notify_recurring_reminders"])
	assert.Equal(t, true, resp["notify_weekly_summary"])
	assert.Equal(t, false, resp["notify_goal_milestones"])
}

func TestSettingsHandler_GET_DefaultThemeAndHideAmounts(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:       1,
		Language: domain.LangEN,
	}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "system", resp["theme"], "missing theme should default to 'system'")
	assert.Equal(t, false, resp["hide_amounts"])
}

func TestSettingsHandler_GET_IncludesThemeAndHideAmounts(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:          1,
		Language:    domain.LangEN,
		Theme:       domain.ThemeDark,
		HideAmounts: true,
	}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "dark", resp["theme"])
	assert.Equal(t, true, resp["hide_amounts"])
}

func TestSettingsHandler_PATCH_Theme(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	existing := &domain.User{ID: 1, Language: domain.LangEN, Theme: domain.ThemeSystem}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
	updated := &domain.User{ID: 1, Language: domain.LangEN, Theme: domain.ThemeDark}
	userRepo.On("UpdateTheme", mock.Anything, int64(1), domain.ThemeDark).Return(updated, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	body := `{"theme":"dark"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "dark", resp["theme"])
	userRepo.AssertExpectations(t)
}

func TestSettingsHandler_PATCH_InvalidTheme(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1}, nil)
	accountRepo := &mocks.MockAccountStorer{}

	h := buildSettingsHandler(userRepo, accountRepo)
	body := `{"theme":"midnight"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	userRepo.AssertNotCalled(t, "UpdateTheme", mock.Anything, mock.Anything, mock.Anything)
}

func TestSettingsHandler_PATCH_HideAmounts(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	existing := &domain.User{ID: 1, Language: domain.LangEN, HideAmounts: false}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)
	updated := &domain.User{ID: 1, Language: domain.LangEN, HideAmounts: true}
	userRepo.On("UpdateHideAmounts", mock.Anything, int64(1), true).Return(updated, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	body := `{"hide_amounts":true}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, true, resp["hide_amounts"])
	userRepo.AssertExpectations(t)
}

func TestSettingsHandler_PATCH_NotificationPrefs(t *testing.T) {
	userRepo := &mocks.MockUserStorer{}
	existing := &domain.User{
		ID:                       1,
		Language:                 domain.LangEN,
		NotifyBudgetAlerts:       true,
		NotifyRecurringReminders: false,
		NotifyWeeklySummary:      false,
		NotifyGoalMilestones:     false,
	}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(existing, nil)

	updatedPrefs := domain.NotificationPrefs{
		BudgetAlerts:       false,
		RecurringReminders: false,
		WeeklySummary:      false,
		GoalMilestones:     false,
	}
	updated := &domain.User{
		ID:                 1,
		Language:           domain.LangEN,
		NotifyBudgetAlerts: false,
	}
	userRepo.On("UpdateNotificationPreferences", mock.Anything, int64(1), updatedPrefs).Return(updated, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(userRepo, accountRepo)
	body := `{"notification_preferences":{"notify_budget_alerts":false}}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, false, resp["notify_budget_alerts"])
	userRepo.AssertExpectations(t)
}
