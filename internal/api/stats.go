package api

import (
	"log/slog"
	"net/http"

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
func statsHandler(statsSvc *service.StatsService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		period := r.URL.Query().Get("period")
		if period == "" {
			period = "month"
		}

		from, to, err := service.PeriodRange(period)
		if err != nil {
			writeError(w, log, err)
			return
		}

		stats, err := statsSvc.ByCategory(ctx, userID, from, to)
		if err != nil {
			writeError(w, log, err)
			return
		}

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

		writeJSON(w, http.StatusOK, statsResponse{Period: period, Items: items})
	}
}
