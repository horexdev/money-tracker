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

func buildCatHandler(repo *mocks.MockCategoryStorer) http.HandlerFunc {
	catSvc := service.NewCategoryService(repo, testutil.TestLogger())
	return api.CategoriesHandlerForTest(catSvc, testutil.TestLogger())
}

func TestCategoriesHandler_GET_All(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	repo.On("ListForUser", mock.Anything, int64(1)).Return([]*domain.Category{
		{ID: 1, UserID: 0, Name: "Food"},
		{ID: 2, UserID: 1, Name: "Salary"},
	}, nil)

	h := buildCatHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	cats := resp["categories"].([]any)
	assert.Len(t, cats, 2)
}

func TestCategoriesHandler_POST_InvalidJSON(t *testing.T) {
	h := buildCatHandler(&mocks.MockCategoryStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString("not-json"))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoriesHandler_POST_EmptyName(t *testing.T) {
	h := buildCatHandler(&mocks.MockCategoryStorer{})
	body := `{"name":"","emoji":"🍔","type":"expense"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCategoriesHandler_POST_Create(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	repo.On("CreateForUser", mock.Anything, int64(1), "Groceries", "🛒", "expense", "#6366f1").
		Return(&domain.Category{ID: 10, UserID: 1, Name: "Groceries"}, nil)

	h := buildCatHandler(repo)
	body := `{"name":"Groceries","emoji":"🛒","type":"expense"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/categories", bytes.NewBufferString(body))
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCategoriesHandler_PUT_SystemCategory(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 0}, nil)

	h := buildCatHandler(repo)
	body := `{"name":"Changed","emoji":"","type":"expense","color":"#fff"}`
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPut, "/api/v1/categories/1", bytes.NewBufferString(body))
	r.URL.Path = "/api/v1/categories/1"
	r = r.WithContext(api.WithUserID(r.Context(), 99))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCategoriesHandler_DELETE_WithTransactions(t *testing.T) {
	repo := &mocks.MockCategoryStorer{}
	repo.On("GetByID", mock.Anything, int64(1)).Return(&domain.Category{ID: 1, UserID: 99}, nil)
	repo.On("CountTransactions", mock.Anything, int64(1)).Return(int64(5), nil)

	h := buildCatHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/1", nil)
	r.URL.Path = "/api/v1/categories/1"
	r = r.WithContext(api.WithUserID(r.Context(), 99))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestCategoriesHandler_DELETE_InvalidID(t *testing.T) {
	h := buildCatHandler(&mocks.MockCategoryStorer{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/abc", nil)
	r.URL.Path = "/api/v1/categories/abc"
	r = r.WithContext(api.WithUserID(r.Context(), 1))
	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
