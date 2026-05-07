package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func buildBalanceHandler(
	txRepo *mocks.MockTransactionStorer,
	userRepo *mocks.MockUserStorer,
	accountRepo *mocks.MockAccountStorer,
) http.HandlerFunc {
	log := testutil.TestLogger()
	userSvc := service.NewUserService(userRepo, log)
	accountSvc := service.NewAccountService(accountRepo, nil, log)
	// exchangeSvc is nil — only used when user.DisplayCurrencies is non-empty.
	return api.BalanceHandlerForTest(txRepo, userSvc, accountSvc, nil, log)
}

func TestBalanceHandler_GET_AggregatesByCurrency(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("GetBalanceByCurrency", mock.Anything, int64(1)).Return([]domain.BalanceByCurrency{
		{CurrencyCode: "USD", IncomeCents: 100000, ExpenseCents: 30000},
		{CurrencyCode: "EUR", IncomeCents: 50000, ExpenseCents: 20000},
	}, nil)
	txRepo.On("GetTotalInBaseCurrency", mock.Anything, int64(1), "USD").Return(int64(70000), nil)

	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "en", DisplayCurrencies: nil}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD", IsDefault: true}, nil)

	h := buildBalanceHandler(txRepo, userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/balance", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	byCurrency := resp["by_currency"].([]any)
	require.Len(t, byCurrency, 2)
	assert.Equal(t, float64(70000), resp["total_in_base_cents"])
}

func TestBalanceHandler_GET_FiltersByAccountID(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("GetBalanceByCurrencyAndAccount", mock.Anything, int64(1), int64(7)).Return([]domain.BalanceByCurrency{
		{CurrencyCode: "USD", IncomeCents: 1000, ExpenseCents: 500},
	}, nil)
	txRepo.On("GetTotalInBaseCurrency", mock.Anything, int64(1), "USD").Return(int64(500), nil)

	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "en"}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(&domain.Account{ID: 1, CurrencyCode: "USD"}, nil)

	h := buildBalanceHandler(txRepo, userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/balance?account_id=7", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	txRepo.AssertCalled(t, "GetBalanceByCurrencyAndAccount", mock.Anything, int64(1), int64(7))
	txRepo.AssertNotCalled(t, "GetBalanceByCurrency", mock.Anything, mock.Anything)
}

func TestBalanceHandler_NonGET_Returns405(t *testing.T) {
	h := buildBalanceHandler(&mocks.MockTransactionStorer{}, &mocks.MockUserStorer{}, &mocks.MockAccountStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/balance", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestBalanceHandler_FallsBackToUSD_WhenNoDefaultAccount(t *testing.T) {
	txRepo := &mocks.MockTransactionStorer{}
	txRepo.On("GetBalanceByCurrency", mock.Anything, int64(1)).Return([]domain.BalanceByCurrency{}, nil)
	txRepo.On("GetTotalInBaseCurrency", mock.Anything, int64(1), "USD").Return(int64(0), nil)

	userRepo := &mocks.MockUserStorer{}
	userRepo.On("GetByID", mock.Anything, int64(1)).Return(&domain.User{ID: 1, Language: "en"}, nil)

	accountRepo := &mocks.MockAccountStorer{}
	accountRepo.On("GetDefault", mock.Anything, int64(1)).Return(nil, domain.ErrAccountNotFound)

	h := buildBalanceHandler(txRepo, userRepo, accountRepo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/balance", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	txRepo.AssertCalled(t, "GetTotalInBaseCurrency", mock.Anything, int64(1), "USD")
}
