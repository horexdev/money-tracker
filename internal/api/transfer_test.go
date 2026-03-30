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

func buildTransferHandler(transferRepo *mocks.MockTransferStorer, accountRepo *mocks.MockAccountStorer, goalRepo *mocks.MockSavingsGoalStorer) http.HandlerFunc {
	svc := service.NewTransferService(transferRepo, accountRepo, goalRepo, testutil.TestLogger())
	return api.TransfersHandlerForTest(svc, testutil.TestLogger())
}

func TestTransferHandler_GET_List(t *testing.T) {
	repo := &mocks.MockTransferStorer{}
	repo.On("Count", mock.Anything, int64(1)).Return(int64(1), nil)
	repo.On("ListByUser", mock.Anything, int64(1), 50, 0).Return([]*domain.Transfer{
		{ID: 1, FromAccountID: 1, ToAccountID: 2, AmountCents: 5000},
	}, nil)

	h := buildTransferHandler(repo, &mocks.MockAccountStorer{}, &mocks.MockSavingsGoalStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/transfers", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	transfers := resp["transfers"].([]any)
	assert.Len(t, transfers, 1)
}

func TestTransferHandler_POST_SameAccount(t *testing.T) {
	h := buildTransferHandler(&mocks.MockTransferStorer{}, &mocks.MockAccountStorer{}, &mocks.MockSavingsGoalStorer{})
	body := `{"from_account_id":1,"to_account_id":1,"amount_cents":1000,"exchange_rate":1.0}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/transfers", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTransferHandler_DELETE_Success(t *testing.T) {
	repo := &mocks.MockTransferStorer{}
	repo.On("Delete", mock.Anything, int64(5), int64(1)).Return(nil)

	h := buildTransferHandler(repo, &mocks.MockAccountStorer{}, &mocks.MockSavingsGoalStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/transfers/5", nil)
	r.URL.Path = "/api/v1/transfers/5"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
