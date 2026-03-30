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

func buildBudgetHandler(budgetRepo *mocks.MockBudgetStorer, txRepo *mocks.MockTransactionStorer) http.HandlerFunc {
	svc := service.NewBudgetService(budgetRepo, txRepo, testutil.TestLogger())
	return api.BudgetHandlerForTest(svc, testutil.TestLogger())
}

func TestBudgetHandler_GET_List(t *testing.T) {
	budgetRepo := &mocks.MockBudgetStorer{}
	budgetRepo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Budget{
		{ID: 1, UserID: 1, LimitCents: 10000, Period: "monthly", CurrencyCode: "USD"},
	}, nil)
	budgetRepo.On("GetSpentInPeriod", mock.Anything, int64(1), int64(0), "USD", mock.Anything, mock.Anything).Return(int64(5000), nil)

	h := buildBudgetHandler(budgetRepo, &mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/budgets", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	budgets := resp["budgets"].([]any)
	assert.Len(t, budgets, 1)
}

func TestBudgetHandler_POST_InvalidJSON(t *testing.T) {
	h := buildBudgetHandler(&mocks.MockBudgetStorer{}, &mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/budgets", bytes.NewBufferString("not-json"))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBudgetHandler_POST_DuplicateBudget(t *testing.T) {
	budgetRepo := &mocks.MockBudgetStorer{}
	// Existing budget found → duplicate.
	budgetRepo.On("GetByUserCategoryPeriod", mock.Anything, int64(1), int64(2), "monthly").
		Return(&domain.Budget{ID: 5}, nil)

	h := buildBudgetHandler(budgetRepo, &mocks.MockTransactionStorer{})
	body := `{"category_id":2,"limit_cents":10000,"period":"monthly","currency_code":"USD","notify_at_percent":80}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/budgets", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestBudgetHandler_POST_Success(t *testing.T) {
	budgetRepo := &mocks.MockBudgetStorer{}
	budgetRepo.On("GetByUserCategoryPeriod", mock.Anything, int64(1), int64(2), "monthly").
		Return(nil, domain.ErrBudgetNotFound)
	budgetRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Budget")).
		Return(&domain.Budget{ID: 10, LimitCents: 10000, Period: "monthly", CurrencyCode: "USD"}, nil)

	h := buildBudgetHandler(budgetRepo, &mocks.MockTransactionStorer{})
	body := `{"category_id":2,"limit_cents":10000,"period":"monthly","currency_code":"USD","notify_at_percent":80}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/budgets", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestBudgetHandler_DELETE_Success(t *testing.T) {
	budgetRepo := &mocks.MockBudgetStorer{}
	budgetRepo.On("Delete", mock.Anything, int64(3), int64(1)).Return(nil)

	h := buildBudgetHandler(budgetRepo, &mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/budgets/3", nil)
	r.URL.Path = "/api/v1/budgets/3"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
