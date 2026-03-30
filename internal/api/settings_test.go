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

func buildSettingsHandler(repo *mocks.MockUserStorer) http.HandlerFunc {
	userSvc := service.NewUserService(repo, testutil.TestLogger())
	return api.SettingsHandlerForTest(userSvc, 0, testutil.TestLogger())
}

func TestSettingsHandler_GET(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{
		ID:                1,
		CurrencyCode:      "USD",
		Language:          domain.LangEN,
		DisplayCurrencies: []string{"EUR"},
	}, nil)

	h := buildSettingsHandler(repo)
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

func TestSettingsHandler_PATCH_InvalidCurrency(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(repo)
	body := `{"base_currency":"INVALID"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSettingsHandler_PATCH_TooManyDisplayCurrencies(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildSettingsHandler(repo)
	body := `{"display_currencies":["USD","EUR","GBP","JPY"]}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSettingsHandler_PATCH_ValidCurrency(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, CurrencyCode: "USD"}, nil)
	repo.On("UpdateCurrency", mock.Anything, int64(1), "EUR").Return(&domain.User{ID: 1, CurrencyCode: "EUR"}, nil)

	h := buildSettingsHandler(repo)
	body := `{"base_currency":"EUR"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/settings", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "EUR", resp["base_currency"])
}

func TestSettingsHandler_UnsupportedMethod(t *testing.T) {
	h := buildSettingsHandler(&mocks.MockUserStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/settings", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
