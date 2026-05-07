package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// buildTxHandler builds a transactionHandler for testing with mocked repos.
func buildTxHandler(txRepo *mocks.MockTransactionStorer, catRepo *mocks.MockCategoryStorer, accountRepo *mocks.MockAccountStorer) http.HandlerFunc {
	log := testutil.TestLogger()
	txSvc := service.NewTransactionService(txRepo, catRepo, log)
	accountSvc := service.NewAccountService(accountRepo, nil, log)
	return api.TransactionHandlerForTest(txSvc, accountSvc, log)
}

func TestTransactionHandler_GET_List(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("Count", mock.Anything, int64(1)).Return(int64(2), nil)
	txRepo.On("List", mock.Anything, int64(1), 20, 0).Return([]*domain.Transaction{
		{ID: 1, AmountCents: 1000, Type: domain.TransactionTypeExpense, AccountID: 1},
		{ID: 2, AmountCents: 2000, Type: domain.TransactionTypeIncome, AccountID: 1},
	}, nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, float64(1), resp["total_pages"])
	txs := resp["transactions"].([]any)
	assert.Len(t, txs, 2)
}

func TestTransactionHandler_POST_InvalidJSON(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString("not-json"))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_ZeroAmount(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	body := `{"type":"expense","amount_cents":0,"category_id":1,"account_id":1}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_MissingCategoryID(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	body := `{"type":"expense","amount_cents":1000,"category_id":0,"account_id":1}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_MissingAccountID(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	body := `{"type":"expense","amount_cents":1000,"category_id":1}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_InvalidType(t *testing.T) {
	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)
	accountRepo.On("GetBalance", mock.Anything, int64(1), int64(1)).Return(int64(0), nil)
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, accountRepo)
	body := `{"type":"transfer","amount_cents":1000,"category_id":1,"account_id":1}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_InvalidDateFormat(t *testing.T) {
	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)
	accountRepo.On("GetBalance", mock.Anything, int64(1), int64(1)).Return(int64(0), nil)
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, accountRepo)
	createdAt := "15/01/2024" // wrong format
	body, _ := json.Marshal(map[string]any{
		"type":         "expense",
		"amount_cents": 1000,
		"category_id":  1,
		"account_id":   1,
		"created_at":   createdAt,
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBuffer(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_POST_ExpenseSuccess(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	catRepo := &mocks.MockCategoryStorer{}
	accountRepo := &mocks.MockAccountStorer{}

	accountRepo.On("GetByID", mock.Anything, int64(1), int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)
	accountRepo.On("GetBalance", mock.Anything, int64(1), int64(1)).Return(int64(0), nil)
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)
	catRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0, Name: "Food"}, nil)
	txRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Transaction")).Return(&domain.Transaction{
		ID: 10, Type: domain.TransactionTypeExpense, AmountCents: 1000, AccountID: 1,
	}, nil)

	h := buildTxHandler(txRepo, catRepo, accountRepo)
	body := `{"type":"expense","amount_cents":1000,"category_id":1,"account_id":1}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transactions", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestTransactionHandler_DELETE_InvalidID(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/abc", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	r.URL.Path = "/api/v1/transactions/abc"

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_DELETE_NotFound(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("Delete", mock.Anything, int64(99), int64(1)).Return(domain.ErrTransactionNotFound)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/99", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	r.URL.Path = "/api/v1/transactions/99"

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTransactionHandler_DELETE_Success(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("Delete", mock.Anything, int64(5), int64(1)).Return(nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/transactions/5", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	r.URL.Path = "/api/v1/transactions/5"

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestTransactionHandler_GET_DateRangeParams(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("CountWithDateRange", mock.Anything, int64(1), mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time")).Return(int64(1), nil)
	txRepo.On("ListWithDateRange", mock.Anything, int64(1), mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time"), 20, 0).Return([]*domain.Transaction{
		{ID: 1, AmountCents: 500, Type: domain.TransactionTypeExpense, AccountID: 1},
	}, nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?from=2025-01-01&to=2025-01-31", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	txs := resp["transactions"].([]any)
	assert.Len(t, txs, 1)
	txRepo.AssertExpectations(t)
}

func TestTransactionHandler_GET_InvalidDateParam(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?from=not-a-date", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransactionHandler_GET_CategoryFilter_Routes(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("CountByCategoryWithDateRange", mock.Anything, int64(1), int64(7),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time")).Return(int64(1), nil)
	txRepo.On("ListByCategoryWithDateRange", mock.Anything, int64(1), int64(7),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time"), 20, 0).
		Return([]*domain.Transaction{
			{ID: 11, AmountCents: 1200, Type: domain.TransactionTypeExpense, CategoryID: 7, AccountID: 1},
		}, nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?category_id=7", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	txs := resp["transactions"].([]any)
	assert.Len(t, txs, 1)
	txRepo.AssertExpectations(t)
}

func TestTransactionHandler_GET_CategoryAndDateRange(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("CountByCategoryWithDateRange", mock.Anything, int64(1), int64(9),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time")).Return(int64(0), nil).
		Run(func(args mock.Arguments) {
			from := args.Get(3).(*time.Time)
			to := args.Get(4).(*time.Time)
			require.NotNil(t, from)
			require.NotNil(t, to)
			require.Equal(t, "2026-04-01", from.Format("2006-01-02"))
			require.Equal(t, "2026-04-30", to.Format("2006-01-02"))
		})
	txRepo.On("ListByCategoryWithDateRange", mock.Anything, int64(1), int64(9),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time"), 20, 0).
		Return([]*domain.Transaction{}, nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?category_id=9&from=2026-04-01&to=2026-04-30", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	txRepo.AssertExpectations(t)
}

func TestTransactionHandler_GET_CategoryAndAccount(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("CountByAccountAndCategoryWithDateRange", mock.Anything, int64(1), int64(3), int64(7),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time")).Return(int64(0), nil)
	txRepo.On("ListByAccountAndCategoryWithDateRange", mock.Anything, int64(1), int64(3), int64(7),
		mock.AnythingOfType("*time.Time"), mock.AnythingOfType("*time.Time"), 20, 0).
		Return([]*domain.Transaction{}, nil)

	h := buildTxHandler(txRepo, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?category_id=7&account_id=3", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	txRepo.AssertExpectations(t)
}

func TestTransactionHandler_GET_InvalidCategoryParam(t *testing.T) {
	h := buildTxHandler(&mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transactions?category_id=abc", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
