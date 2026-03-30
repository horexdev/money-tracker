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

func buildRecurringHandler(recurringRepo *mocks.MockRecurringStorer, txRepo *mocks.MockTransactionStorer) http.HandlerFunc {
	svc := service.NewRecurringService(recurringRepo, txRepo, testutil.TestLogger())
	return api.RecurringHandlerForTest(svc, testutil.TestLogger())
}

func TestRecurringHandler_GET_List(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.RecurringTransaction{
		{ID: 1, AmountCents: 5000, Frequency: "monthly", IsActive: true},
	}, nil)

	h := buildRecurringHandler(repo, &mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/recurring", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	items := resp["recurring"].([]any)
	assert.Len(t, items, 1)
}

func TestRecurringHandler_POST_ZeroAmount(t *testing.T) {
	h := buildRecurringHandler(&mocks.MockRecurringStorer{}, &mocks.MockTransactionStorer{})
	body := `{"amount_cents":0,"frequency":"monthly","type":"expense","category_id":1,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/recurring", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRecurringHandler_POST_InvalidFrequency(t *testing.T) {
	h := buildRecurringHandler(&mocks.MockRecurringStorer{}, &mocks.MockTransactionStorer{})
	body := `{"amount_cents":1000,"frequency":"biweekly","type":"expense","category_id":1,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/recurring", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRecurringHandler_DELETE_Success(t *testing.T) {
	repo := &mocks.MockRecurringStorer{}
	repo.On("Delete", mock.Anything, int64(3), int64(1)).Return(nil)

	h := buildRecurringHandler(repo, &mocks.MockTransactionStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/recurring/3", nil)
	r.URL.Path = "/api/v1/recurring/3"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
