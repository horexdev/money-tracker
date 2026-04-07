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

func buildAccountHandler(repo *mocks.MockAccountStorer) http.HandlerFunc {
	// nil ExchangeService — tests avoid balance conversion paths.
	accountSvc := service.NewAccountService(repo, nil, testutil.TestLogger())
	// nil repositories for AdjustmentService — adjust endpoint not exercised in these tests.
	adjustSvc := service.NewAdjustmentService(&mocks.MockTransactionStorer{}, repo, &mocks.MockCategoryStorer{}, testutil.TestLogger())
	return api.AccountsHandlerForTest(accountSvc, adjustSvc, testutil.TestLogger())
}

func TestAccountsHandler_GET_List(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Account{
		{ID: 1, Name: "Main", IsDefault: true, CurrencyCode: "USD"},
	}, nil)
	// balanceInCurrency path: GetBalanceInBase returns 0 → no further calls.
	repo.On("GetBalanceInBase", mock.Anything, int64(1), int64(1)).Return(int64(0), nil)

	h := buildAccountHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	accounts := resp["accounts"].([]any)
	assert.Len(t, accounts, 1)
}

func TestAccountsHandler_POST_MissingName(t *testing.T) {
	h := buildAccountHandler(&mocks.MockAccountStorer{})
	body := `{"icon":"wallet","currency_code":"USD","type":"checking"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAccountsHandler_POST_MissingCurrency(t *testing.T) {
	h := buildAccountHandler(&mocks.MockAccountStorer{})
	body := `{"name":"Main","icon":"wallet","type":"checking"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAccountsHandler_POST_Create(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.Account{}, nil)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Account")).Return(&domain.Account{
		ID: 1, Name: "Main", IsDefault: true, CurrencyCode: "USD",
	}, nil)

	h := buildAccountHandler(repo)
	b := true
	body, _ := json.Marshal(map[string]any{
		"name":             "Main",
		"icon":             "wallet",
		"color":            "#fff",
		"type":             "checking",
		"currency_code":    "USD",
		"include_in_total": b,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBuffer(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAccountsHandler_DELETE_HasTransactions(t *testing.T) {
	repo := &mocks.MockAccountStorer{}
	repo.On("CountTransactions", mock.Anything, int64(5), int64(1)).Return(int64(3), nil)

	h := buildAccountHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/5", nil)
	r.URL.Path = "/api/v1/accounts/5"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusConflict, w.Code)
}
