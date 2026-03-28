package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const contextKeyUserID contextKey = "userID"

// userIDFromContext retrieves the authenticated user ID from the context.
func userIDFromContext(ctx context.Context) int64 {
	id, _ := ctx.Value(contextKeyUserID).(int64)
	return id
}

// ensurer is a minimal interface for ensuring a user exists before handling a request.
type ensurer interface {
	ensureUser(ctx context.Context, userID int64) error
}

// authMiddleware validates the Telegram initData header and injects userID into the context.
// It also upserts the user on first contact, reusing the existing UserService.
func authMiddleware(botToken string, userSvc ensurer, log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initData := r.Header.Get("X-Telegram-Init-Data")
			userID, err := ValidateInitData(botToken, initData)
			if err != nil {
				if errors.Is(err, ErrInvalidInitData) {
					writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
					return
				}
				log.ErrorContext(r.Context(), "auth: unexpected validation error", slog.String("error", err.Error()))
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}

			// Ensure the user exists in the database.
			if err := userSvc.ensureUser(r.Context(), userID); err != nil {
				log.ErrorContext(r.Context(), "auth: failed to ensure user", slog.Int64("user_id", userID), slog.String("error", err.Error()))
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// corsMiddleware adds CORS headers. allowedOrigins is a comma-separated list of
// allowed origins, or "*" to allow all.
func corsMiddleware(allowedOrigins string) func(http.Handler) http.Handler {
	origins := make(map[string]bool)
	allowAll := allowedOrigins == "*"
	if !allowAll {
		for _, o := range strings.Split(allowedOrigins, ",") {
			origins[strings.TrimSpace(o)] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if allowAll || origins[origin] {
				w.Header().Set("Access-Control-Allow-Origin", func() string {
					if allowAll {
						return "*"
					}
					return origin
				}())
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Telegram-Init-Data")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// loggingMiddleware logs each request with method, path, status, and duration.
func loggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.InfoContext(r.Context(), "api request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// chain applies a list of middleware in order (outermost first).
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
