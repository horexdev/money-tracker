package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/horexdev/money-tracker/internal/testutil/mocks"
)

func buildUserDataHandler(repo *mocks.MockUserStorer) http.HandlerFunc {
	log := testutil.TestLogger()
	userSvc := service.NewUserService(repo, log)
	return api.UserDataHandlerForTest(userSvc, log)
}

func newAuthedUserRequest(method string) *http.Request {
	r := httptest.NewRequest(method, "/api/v1/user/data", nil)
	return r.WithContext(api.WithUserID(context.Background(), 1))
}

func TestUserDataHandler_DELETE_Success(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("ResetData", mock.Anything, int64(1)).Return(nil)

	h := buildUserDataHandler(repo)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAuthedUserRequest(http.MethodDelete))

	assert.Equal(t, http.StatusNoContent, w.Code)
	repo.AssertExpectations(t)
}

func TestUserDataHandler_NonDELETE_Returns405(t *testing.T) {
	h := buildUserDataHandler(&mocks.MockUserStorer{})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAuthedUserRequest(http.MethodGet))

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUserDataHandler_PropagatesServiceError(t *testing.T) {
	repo := &mocks.MockUserStorer{}
	repo.On("ResetData", mock.Anything, int64(1)).Return(errors.New("db down"))

	h := buildUserDataHandler(repo)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, newAuthedUserRequest(http.MethodDelete))

	assert.GreaterOrEqual(t, w.Code, 400)
}
