package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/horexdev/money-tracker/internal/api"
	"github.com/horexdev/money-tracker/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noopEnsure is an EnsureUserFunc that always succeeds.
var noopEnsure api.EnsureUserFunc = func(_ context.Context, _ api.TelegramUser) error { return nil }

// errEnsure is an EnsureUserFunc that always returns an error.
var errEnsure api.EnsureUserFunc = func(_ context.Context, _ api.TelegramUser) error {
	return errors.New("db down")
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	mw := api.AuthMiddlewareForTest(testBotToken, false, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidInitData(t *testing.T) {
	mw := api.AuthMiddlewareForTest(testBotToken, false, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", "garbage=data&hash=abc")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_EnsureUserError(t *testing.T) {
	mw := api.AuthMiddlewareForTest(testBotToken, false, errEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", testutil.BuildValidInitData(testBotToken, 42, "Alice", time.Now()))
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthMiddleware_ValidInitData(t *testing.T) {
	var capturedUserID int64
	mw := api.AuthMiddlewareForTest(testBotToken, false, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", testutil.BuildValidInitData(testBotToken, 99, "Bob", time.Now()))

	next := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedUserID = api.UserIDFromContextForTest(req.Context())
		w.WriteHeader(http.StatusOK)
	})
	mw(next).ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(99), capturedUserID)
}

func TestAdminMiddleware_WrongUser(t *testing.T) {
	mw := api.AdminMiddlewareForTest(1000)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 999))
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminMiddleware_ZeroAdminID_Forbidden(t *testing.T) {
	mw := api.AdminMiddlewareForTest(0)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1000))
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminMiddleware_CorrectUser(t *testing.T) {
	mw := api.AdminMiddlewareForTest(1000)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r = r.WithContext(api.WithUserID(r.Context(), 1000))
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCorsMiddleware_AllowAll(t *testing.T) {
	mw := api.CorsMiddlewareForTest("*")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsMiddleware_AllowedOrigin(t *testing.T) {
	mw := api.CorsMiddlewareForTest("https://example.com,https://other.com")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsMiddleware_DisallowedOrigin(t *testing.T) {
	mw := api.CorsMiddlewareForTest("https://example.com")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Origin", "https://evil.com")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCorsMiddleware_Preflight(t *testing.T) {
	mw := api.CorsMiddlewareForTest("*")
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Set("Origin", "https://example.com")
	var nextCalled bool
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true })
	mw(next).ServeHTTP(w, r)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.False(t, nextCalled, "next handler should not be called for OPTIONS preflight")
}

func TestAuthMiddleware_DevMode_ValidDevToken(t *testing.T) {
	var capturedUserID int64
	mw := api.AuthMiddlewareForTest(testBotToken, true, noopEnsure)
	next := http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		capturedUserID = api.UserIDFromContextForTest(req.Context())
	})
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", "dev:12345")
	mw(next).ServeHTTP(w, r)
	assert.Equal(t, int64(12345), capturedUserID)
}

func TestAuthMiddleware_DevMode_InvalidDevToken(t *testing.T) {
	mw := api.AuthMiddlewareForTest(testBotToken, true, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", "dev:notanumber")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_DevMode_ZeroUserID(t *testing.T) {
	mw := api.AuthMiddlewareForTest(testBotToken, true, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", "dev:0")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_DevModeDisabled_DevTokenRejected(t *testing.T) {
	// When devMode is false, dev: prefix should NOT bypass — normal HMAC validation occurs.
	mw := api.AuthMiddlewareForTest(testBotToken, false, noopEnsure)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Telegram-Init-Data", "dev:12345")
	mw(okHandler()).ServeHTTP(w, r)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// okHandler returns an HTTP handler that always responds 200 OK.
func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
