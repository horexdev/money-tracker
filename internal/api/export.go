package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/horexdev/money-tracker/internal/service"
)

// exportHandler handles GET /api/v1/export?format=csv&from=2006-01-02&to=2006-01-02
func exportHandler(svc *service.ExportService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		userID := userIDFromContext(ctx)
		q := r.URL.Query()

		format := q.Get("format")
		if format == "" {
			format = "csv"
		}
		if format != "csv" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unsupported format, use csv"})
			return
		}

		fromStr := q.Get("from")
		toStr := q.Get("to")
		if fromStr == "" || toStr == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "from and to query parameters are required (YYYY-MM-DD)"})
			return
		}

		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid from date, use YYYY-MM-DD"})
			return
		}
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid to date, use YYYY-MM-DD"})
			return
		}
		// Include the full "to" day.
		to = to.AddDate(0, 0, 1)

		data, err := svc.ExportCSV(ctx, userID, from, to)
		if err != nil {
			writeError(w, log, err)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
