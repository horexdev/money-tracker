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

func buildGoalsHandler(repo *mocks.MockSavingsGoalStorer) http.HandlerFunc {
	svc := service.NewSavingsGoalService(repo, &mocks.MockTransactionStorer{}, &mocks.MockCategoryStorer{}, &mocks.MockAccountStorer{}, testutil.TestLogger())
	return api.GoalsHandlerForTest(svc, testutil.TestLogger())
}

func TestGoalsHandler_GET_List(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	repo.On("ListByUser", mock.Anything, int64(1)).Return([]*domain.SavingsGoal{
		{ID: 1, Name: "Vacation", TargetCents: 100000, CurrentCents: 50000},
	}, nil)

	h := buildGoalsHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/goals", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	goals := resp["goals"].([]any)
	assert.Len(t, goals, 1)
}

func TestGoalsHandler_POST_ZeroTarget(t *testing.T) {
	h := buildGoalsHandler(&mocks.MockSavingsGoalStorer{})
	body := `{"name":"Vacation","target_cents":0,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/goals", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGoalsHandler_POST_Create(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.SavingsGoal")).
		Return(&domain.SavingsGoal{ID: 1, Name: "Vacation", TargetCents: 50000, CurrencyCode: "USD"}, nil)

	h := buildGoalsHandler(repo)
	body := `{"name":"Vacation","target_cents":50000,"currency_code":"USD"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/goals", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestGoalsHandler_DELETE_Success(t *testing.T) {
	repo := &mocks.MockSavingsGoalStorer{}
	repo.On("Delete", mock.Anything, int64(7), int64(1)).Return(nil)

	h := buildGoalsHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/goals/7", nil)
	r.URL.Path = "/api/v1/goals/7"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
}
