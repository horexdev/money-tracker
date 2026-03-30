package api

import (
	"log/slog"
	"net/http"
	"strconv"
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
