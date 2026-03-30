package api

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/horexdev/money-tracker/internal/service"
)

type adminUserJSON struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	CurrencyCode string `json:"currency_code"`
	Language     string `json:"language"`
	CreatedAt    string `json:"created_at"`
}

type adminUsersResponse struct {
	Users    []adminUserJSON `json:"users"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

// adminUsersHandler handles GET /api/v1/admin/users.
func adminUsersHandler(svc *service.AdminService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
		if pageSize < 1 || pageSize > 100 {
			pageSize = 20
		}

		users, total, err := svc.ListUsers(r.Context(), page, pageSize)
		if err != nil {
			writeError(w, log, err)
			return
		}

		items := make([]adminUserJSON, 0, len(users))
		for _, u := range users {
			items = append(items, adminUserJSON{
				ID:           u.ID,
				Username:     u.Username,
				FirstName:    u.FirstName,
				LastName:     u.LastName,
				CurrencyCode: u.CurrencyCode,
				Language:     string(u.Language),
				CreatedAt:    u.CreatedAt.Format(time.RFC3339),
			})
		}

		writeJSON(w, http.StatusOK, adminUsersResponse{
			Users:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

// adminStatsHandler handles GET /api/v1/admin/stats.
func adminStatsHandler(svc *service.AdminService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		stats, err := svc.GetStats(r.Context())
		if err != nil {
			writeError(w, log, err)
			return
		}

		writeJSON(w, http.StatusOK, stats)
	}
}

// adminResetUserHandler handles DELETE /api/v1/admin/users/{id}/data.
// Resets all data for the specified user and recreates their default account.
func adminResetUserHandler(adminSvc *service.AdminService, userSvc *service.UserService, accountSvc *service.AccountService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Extract {id} from /api/v1/admin/users/{id}/data
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/admin/users/")
		path = strings.TrimSuffix(path, "/data")
		targetID, err := strconv.ParseInt(path, 10, 64)
		if err != nil || targetID <= 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
			return
		}

		ctx := r.Context()
		if err := resetAndRecreate(ctx, targetID, userSvc, accountSvc, log); err != nil {
			writeError(w, log, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"reset": true, "user_id": targetID})
	}
}

// adminResetAllHandler handles DELETE /api/v1/admin/users/data.
// Resets all data for every registered user and recreates their default accounts.
func adminResetAllHandler(adminSvc *service.AdminService, userSvc *service.UserService, accountSvc *service.AccountService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		ids, err := adminSvc.ListAllUserIDs(ctx)
		if err != nil {
			writeError(w, log, err)
			return
		}

		var failed []int64
		for _, id := range ids {
			if err := resetAndRecreate(ctx, id, userSvc, accountSvc, log); err != nil {
				log.ErrorContext(ctx, "admin: reset user failed",
					slog.Int64("user_id", id),
					slog.String("error", err.Error()),
				)
				failed = append(failed, id)
			}
		}

		writeJSON(w, http.StatusOK, map[string]any{
			"reset":  len(ids) - len(failed),
			"failed": len(failed),
		})
	}
}

// resetAndRecreate wipes all data for targetID including the user record itself.
// The user will be fully re-initialised (language, currency, default account) on next login.
func resetAndRecreate(ctx context.Context, targetID int64, userSvc *service.UserService, _ *service.AccountService, _ *slog.Logger) error {
	return userSvc.ResetData(ctx, targetID)
}
