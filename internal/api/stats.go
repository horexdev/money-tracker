package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type categoryStatResponse struct {
	CategoryName  string `json:"category_name"`
	CategoryEmoji string `json:"category_emoji"`
	CategoryColor string `json:"category_color"`
	Type          string `json:"type"`
	TotalCents    int64  `json:"total_cents"`
	TxCount       int64  `json:"tx_count"`
	CurrencyCode  string `json:"currency_code"`
}

type statsResponse struct {
	Period string                 `json:"period"`
	Items  []categoryStatResponse `json:"items"`
}

// statsHandler handles GET /api/v1/stats?period=month
// or GET /api/v1/stats?from=2024-01-01&to=2024-01-31 for a custom date range.
func statsHandler(statsSvc *service.StatsService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		q := r.URL.Query()
		fromStr := q.Get("from")
		toStr := q.Get("to")

		var accountID int64
		if accountIDStr := q.Get("account_id"); accountIDStr != "" {
			if id, parseErr := strconv.ParseInt(accountIDStr, 10, 64); parseErr == nil && id > 0 {
				accountID = id
			}
		}

		fetchStats := func(from, to time.Time) ([]domain.CategoryStat, error) {
			if accountID > 0 {
				return statsSvc.ByCategoryAndAccount(ctx, userID, accountID, from, to)
			}
			return statsSvc.ByCategory(ctx, userID, from, to)
		}

		if fromStr != "" && toStr != "" {
			// Custom range: parse ISO date strings (YYYY-MM-DD)
			f, err := time.Parse("2006-01-02", fromStr)
			if err != nil {
				http.Error(w, "invalid from date", http.StatusBadRequest)
				return
			}
			t, err := time.Parse("2006-01-02", toStr)
			if err != nil {
				http.Error(w, "invalid to date", http.StatusBadRequest)
				return
			}
			// to is inclusive: advance by one day to capture the full end date
			stats, err := fetchStats(f, t.AddDate(0, 0, 1))
			if err != nil {
				writeError(w, log, err)
				return
			}
			writeJSON(w, http.StatusOK, statsResponse{Period: "custom", Items: buildItems(stats)})
			return
		}

		// Named period
		period := q.Get("period")
		if period == "" {
			period = "month"
		}

		fromTime, toTime, err := service.PeriodRange(period)
		if err != nil {
			writeError(w, log, err)
			return
		}

		stats, err := fetchStats(fromTime, toTime)
		if err != nil {
			writeError(w, log, err)
			return
		}

		writeJSON(w, http.StatusOK, statsResponse{Period: period, Items: buildItems(stats)})
	}
}

func buildItems(stats []domain.CategoryStat) []categoryStatResponse {
	items := make([]categoryStatResponse, 0, len(stats))
	for _, s := range stats {
		items = append(items, categoryStatResponse{
			CategoryName:  s.CategoryName,
			CategoryEmoji: s.CategoryEmoji,
			CategoryColor: s.CategoryColor,
			Type:          string(s.Type),
			TotalCents:    s.TotalCents,
			TxCount:       s.TxCount,
			CurrencyCode:  s.CurrencyCode,
		})
	}
	return items
}
