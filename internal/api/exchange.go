package api

import (
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/service"
)

// exchangeRateHandler handles GET /api/v1/exchange/rate?from=USD&to=RUB.
func exchangeRateHandler(svc *service.ExchangeService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		if from == "" || to == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "from and to query params are required"})
			return
		}

		if from == to {
			writeJSON(w, http.StatusOK, map[string]any{"rate": 1.0})
			return
		}

		rate, err := svc.GetRate(r.Context(), from, to)
		if err != nil {
			writeError(w, log, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]any{"rate": rate})
	}
}
