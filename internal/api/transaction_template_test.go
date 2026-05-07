package api_test

import (
	"bytes"
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

func buildTemplateHandler(repo *mocks.MockTransactionTemplateStorer, txAdder *mocks.MockTransactionAdder) http.HandlerFunc {
	svc := service.NewTransactionTemplateService(repo, txAdder, &mocks.MockAccountStorer{}, testutil.TestLogger())
	return api.TemplateHandlerForTest(svc, testutil.TestLogger())
}

func TestTemplateHandler_GET_List(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.TransactionTemplate{
		{ID: 1, Name: "Coffee", Type: domain.TransactionTypeExpense, AmountCents: 30000, AmountFixed: true, CurrencyCode: "USD", SortOrder: 0, CreatedAt: time.Now()},
	}, nil)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/templates", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	items := resp["templates"].([]any)
	assert.Len(t, items, 1)
}

func TestTemplateHandler_POST_Created(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TransactionTemplate")).
		Return(&domain.TransactionTemplate{ID: 42, Name: "Coffee", Type: domain.TransactionTypeExpense, AmountCents: 30000, AmountFixed: true, CurrencyCode: "USD"}, nil)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	body := `{"name":"Coffee","type":"expense","amount_cents":30000,"amount_fixed":true,"category_id":2,"account_id":7,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewBufferString(body))
	r.ContentLength = int64(len(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTemplateHandler_POST_ZeroAmount_BadRequest(t *testing.T) {
	h := buildTemplateHandler(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{})
	body := `{"type":"expense","amount_cents":0,"category_id":1,"account_id":1,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_POST_InvalidJSON_BadRequest(t *testing.T) {
	h := buildTemplateHandler(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates", bytes.NewBufferString("not json"))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_DELETE_Success(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("Delete", mock.Anything, int64(3), int64(1)).Return(nil)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/3", nil)
	r.URL.Path = "/api/v1/templates/3"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTemplateHandler_DELETE_NotFound(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("Delete", mock.Anything, int64(404), int64(1)).Return(domain.ErrTemplateNotFound)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/404", nil)
	r.URL.Path = "/api/v1/templates/404"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTemplateHandler_DELETE_BadID(t *testing.T) {
	h := buildTemplateHandler(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/templates/abc", nil)
	r.URL.Path = "/api/v1/templates/abc"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTemplateHandler_POST_Apply_Fixed_Created(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	txAdder := &mocks.MockTransactionAdder{}

	tpl := &domain.TransactionTemplate{
		ID: 5, UserID: 1, Type: domain.TransactionTypeExpense,
		AmountCents: 30000, AmountFixed: true, CurrencyCode: "USD",
		CategoryID: 2, AccountID: 7, Note: "coffee",
	}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(tpl, nil)
	txAdder.On("AddExpense", mock.Anything, int64(1), int64(30000), int64(2), "coffee", "USD", int64(7), (*time.Time)(nil)).
		Return(&domain.Transaction{ID: 100, AmountCents: 30000}, nil)

	h := buildTemplateHandler(repo, txAdder)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates/5/apply", nil)
	r.URL.Path = "/api/v1/templates/5/apply"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTemplateHandler_POST_Apply_Variable_UsesOverride(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	txAdder := &mocks.MockTransactionAdder{}

	tpl := &domain.TransactionTemplate{
		ID: 5, UserID: 1, Type: domain.TransactionTypeExpense,
		AmountCents: 50000, AmountFixed: false, CurrencyCode: "USD",
		CategoryID: 2, AccountID: 7,
	}
	repo.On("GetByID", mock.Anything, int64(5), int64(1)).Return(tpl, nil)
	txAdder.On("AddExpense", mock.Anything, int64(1), int64(75000), int64(2), "", "USD", int64(7), (*time.Time)(nil)).
		Return(&domain.Transaction{ID: 101, AmountCents: 75000}, nil)

	h := buildTemplateHandler(repo, txAdder)
	body := `{"amount_cents":75000}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates/5/apply", bytes.NewBufferString(body))
	r.URL.Path = "/api/v1/templates/5/apply"
	r.ContentLength = int64(len(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTemplateHandler_POST_Apply_NotFound(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("GetByID", mock.Anything, int64(404), int64(1)).Return(nil, domain.ErrTemplateNotFound)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/templates/404/apply", nil)
	r.URL.Path = "/api/v1/templates/404/apply"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTemplateHandler_PATCH_Reorder_OK(t *testing.T) {
	repo := &mocks.MockTransactionTemplateStorer{}
	repo.On("Reorder", mock.Anything, int64(1), []int64{3, 1, 2}).Return(nil)
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.TransactionTemplate{
		{ID: 3, SortOrder: 0}, {ID: 1, SortOrder: 1}, {ID: 2, SortOrder: 2},
	}, nil)

	h := buildTemplateHandler(repo, &mocks.MockTransactionAdder{})
	body := `{"order":[3,1,2]}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/api/v1/templates/reorder", bytes.NewBufferString(body))
	r.URL.Path = "/api/v1/templates/reorder"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Len(t, resp["templates"].([]any), 3)
}

func TestTemplateHandler_GET_MethodNotAllowedOnReorder(t *testing.T) {
	h := buildTemplateHandler(&mocks.MockTransactionTemplateStorer{}, &mocks.MockTransactionAdder{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/templates/reorder", nil)
	r.URL.Path = "/api/v1/templates/reorder"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
