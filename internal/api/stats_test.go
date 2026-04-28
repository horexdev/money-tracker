package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func buildStatsHandler(txRepo *mocks.MockTransactionStorer) http.HandlerFunc {
	log := testutil.TestLogger()
	statsSvc := service.NewStatsService(txRepo, log)
	return api.StatsHandlerForTest(statsSvc, log)
}

func TestStatsHandler_GET_DefaultPeriodIsMonth(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("StatsByCategory", mock.Anything, int64(1), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]domain.CategoryStat{
			{CategoryName: "Food", Type: domain.TransactionTypeExpense, TotalCents: 1500, TxCount: 3, CurrencyCode: "USD"},
		}, nil)

	h := buildStatsHandler(txRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "month", resp["period"])
	assert.Len(t, resp["items"].([]any), 1)
}

func TestStatsHandler_GET_AcceptsCustomDateRange(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("StatsByCategory", mock.Anything, int64(1), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]domain.CategoryStat{}, nil).
		Run(func(args mock.Arguments) {
			from := args.Get(2).(time.Time)
			to := args.Get(3).(time.Time)
			// from is the parsed "2026-04-01"; to is the day-after parsed "2026-04-30"
			require.Equal(t, "2026-04-01", from.Format("2006-01-02"))
			require.Equal(t, "2026-05-01", to.Format("2006-01-02"))
		})

	h := buildStatsHandler(txRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/stats?from=2026-04-01&to=2026-04-30", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "custom", resp["period"])
}

func TestStatsHandler_GET_400OnInvalidFromDate(t *testing.T) {
	h := buildStatsHandler(&mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/stats?from=not-a-date&to=2026-04-30", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandler_GET_400OnInvalidToDate(t *testing.T) {
	h := buildStatsHandler(&mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/stats?from=2026-04-01&to=invalid", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestStatsHandler_GET_RoutesToAccountVariantWhenAccountIDProvided(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("StatsByCategoryAndAccount", mock.Anything, int64(1), int64(7),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).
		Return([]domain.CategoryStat{}, nil)

	h := buildStatsHandler(txRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/stats?period=month&account_id=7", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	txRepo.AssertCalled(t, "StatsByCategoryAndAccount", mock.Anything, int64(1), int64(7),
		mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time"))
}

func TestStatsHandler_NonGET_Returns405(t *testing.T) {
	h := buildStatsHandler(&mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/stats", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
