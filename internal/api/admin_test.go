package api_test

import (
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

func buildAdminUsersHandler(repo *mocks.MockAdminStorer) http.HandlerFunc {
	svc := service.NewAdminService(repo, testutil.TestLogger())
	return api.AdminUsersHandlerForTest(svc, testutil.TestLogger())
}

func TestAdminUsersHandler_ListUsers(t *testing.T) {
	repo := &mocks.MockAdminStorer{}
	repo.On("ListUsers", mock.Anything, 20, 0).Return([]*domain.User{
		{ID: 1, Username: "alice"},
		{ID: 2, Username: "bob"},
	}, nil)
	repo.On("CountUsers", mock.Anything).Return(int64(2), nil)

	h := buildAdminUsersHandler(repo)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 999))
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	users := resp["users"].([]any)
	assert.Len(t, users, 2)
}

func TestAdminMiddleware_RejectsNonAdmin(t *testing.T) {
	// Verify that non-admin user gets 403 when adminMiddleware is applied.
	const adminID int64 = 1000

	mw := api.AdminMiddlewareForTest(adminID)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 999)) // not admin

	mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(w, r)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
