package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type transferResponse struct {
	ID               int64   `json:"id"`
	FromAccountID    int64   `json:"from_account_id"`
	FromAccountName  string  `json:"from_account_name"`
	ToAccountID      int64   `json:"to_account_id"`
	ToAccountName    string  `json:"to_account_name"`
	AmountCents      int64   `json:"amount_cents"`
	FromCurrencyCode string  `json:"from_currency_code"`
	ToCurrencyCode   string  `json:"to_currency_code"`
	ExchangeRate     float64 `json:"exchange_rate"`
	Note             string  `json:"note"`
	CreatedAt        string  `json:"created_at"`
}

type createTransferRequest struct {
	FromAccountID int64   `json:"from_account_id"`
	ToAccountID   int64   `json:"to_account_id"`
	AmountCents   int64   `json:"amount_cents"`
	ExchangeRate  float64 `json:"exchange_rate"`
	Note          string  `json:"note"`
}

// transfersHandler routes requests for /api/v1/transfers[/{id}].
func transfersHandler(svc *service.TransferService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/transfers")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			id, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid transfer id"})
				return
			}

			switch r.Method {
			case http.MethodGet:
				t, err := svc.GetByID(ctx, id, userID)
				if err != nil {
					writeError(w, log, err)
					return
				}
				writeJSON(w, http.StatusOK, transferToResponse(t))
			case http.MethodDelete:
				if err := svc.Delete(ctx, id, userID); err != nil {
					writeError(w, log, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			listTransfers(w, r, userID, svc, log)
		case http.MethodPost:
			createTransfer(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listTransfers(w http.ResponseWriter, r *http.Request, userID int64, svc *service.TransferService, log *slog.Logger) {
	const defaultLimit = 50
	limit := defaultLimit
	offset := 0

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	transfers, err := svc.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, log, err)
		return
	}
	total, err := svc.Count(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]transferResponse, 0, len(transfers))
	for _, t := range transfers {
		items = append(items, transferToResponse(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"transfers": items,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func createTransfer(w http.ResponseWriter, r *http.Request, userID int64, svc *service.TransferService, log *slog.Logger) {
	var req createTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if req.FromAccountID == 0 || req.ToAccountID == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "from_account_id and to_account_id are required"})
		return
	}
	if req.AmountCents <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "amount_cents must be positive"})
		return
	}

	t, err := svc.Execute(r.Context(), userID, req.FromAccountID, req.ToAccountID,
		req.AmountCents, req.ExchangeRate, req.Note)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, transferToResponse(t))
}

func transferToResponse(t *domain.Transfer) transferResponse {
	return transferResponse{
		ID:               t.ID,
		FromAccountID:    t.FromAccountID,
		FromAccountName:  t.FromAccountName,
		ToAccountID:      t.ToAccountID,
		ToAccountName:    t.ToAccountName,
		AmountCents:      t.AmountCents,
		FromCurrencyCode: t.FromCurrencyCode,
		ToCurrencyCode:   t.ToCurrencyCode,
		ExchangeRate:     t.ExchangeRate,
		Note:             t.Note,
		CreatedAt:        t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
