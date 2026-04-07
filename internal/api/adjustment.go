package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
)

type adjustmentApplier interface {
	Apply(ctx context.Context, userID, accountID, deltaCents int64, note string) (*domain.Transaction, error)
}

type adjustAccountRequest struct {
	DeltaCents int64  `json:"delta_cents"`
	Note       string `json:"note"`
}

// adjustAccountHandler handles POST /api/v1/accounts/{id}/adjust.
// It applies a signed balance correction to the account.
// delta_cents may be positive (increase balance) or negative (decrease balance).
func adjustAccountHandler(svc adjustmentApplier, log *slog.Logger, accountID int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req adjustAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
		if req.DeltaCents == 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: domain.ErrAdjustmentZeroAmount.Error()})
			return
		}

		ctx := r.Context()
		userID := userIDFromContext(ctx)

		tx, err := svc.Apply(ctx, userID, accountID, req.DeltaCents, req.Note)
		if err != nil {
			writeError(w, log, err)
			return
		}
		writeJSON(w, http.StatusCreated, txToResponse(tx))
	}
}
