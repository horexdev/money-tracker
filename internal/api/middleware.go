package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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
	ensureUser(ctx context.Context, tgUser TelegramUser) error
}

// authMiddleware validates the Telegram initData header and injects userID into the context.
// It also upserts the user profile (username, first/last name) on every request so that
// profile data is always kept up to date.
//
// When devMode is true, it additionally accepts the header value "dev:<user_id>" as a bypass,
// skipping HMAC validation entirely. This is for local development only and must never be
// enabled in production.
func authMiddleware(botToken string, devMode bool, devLang string, userSvc ensurer, log *slog.Logger) func(http.Handler) http.Handler {
	if devLang == "" {
		devLang = "en"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			initData := r.Header.Get("X-Telegram-Init-Data")

			// DEV bypass: accept "dev:<user_id>" when devMode is enabled.
			if devMode && strings.HasPrefix(initData, "dev:") {
				userIDStr := strings.TrimPrefix(initData, "dev:")
				userID, err := strconv.ParseInt(userIDStr, 10, 64)
				if err != nil || userID <= 0 {
					writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
					return
				}
				tgUser := TelegramUser{ID: userID, FirstName: "DevUser", LanguageCode: devLang}
				if err := userSvc.ensureUser(r.Context(), tgUser); err != nil {
					log.ErrorContext(r.Context(), "auth: dev mode failed to ensure user",
						slog.Int64("user_id", userID), slog.String("error", err.Error()))
					writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
					return
				}
				log.InfoContext(r.Context(), "auth: dev bypass", slog.Int64("user_id", userID))
				ctx := context.WithValue(r.Context(), contextKeyUserID, userID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tgUser, err := ValidateInitData(botToken, initData)
			if err != nil {
				if errors.Is(err, ErrInvalidInitData) {
					writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
					return
				}
				log.ErrorContext(r.Context(), "auth: unexpected validation error", slog.String("error", err.Error()))
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}

			// Upsert user so profile fields (username, names) are always fresh.
			if err := userSvc.ensureUser(r.Context(), tgUser); err != nil {
				log.ErrorContext(r.Context(), "auth: failed to ensure user", slog.Int64("user_id", tgUser.ID), slog.String("error", err.Error()))
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, tgUser.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// adminMiddleware restricts access to the configured admin Telegram user ID.
// Must be applied after authMiddleware so that userID is present in context.
// When devMode is true, all authenticated users are granted admin access.
func adminMiddleware(adminUserID int64, devMode bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devMode {
				next.ServeHTTP(w, r)
				return
			}
			if adminUserID == 0 || userIDFromContext(r.Context()) != adminUserID {
				writeJSON(w, http.StatusForbidden, errorResponse{Error: "forbidden"})
				return
			}
			next.ServeHTTP(w, r)
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
