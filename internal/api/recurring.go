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

type recurringResponse struct {
	ID            int64  `json:"id"`
	AccountID     int64  `json:"account_id"`
	Type          string `json:"type"`
	AmountCents   int64  `json:"amount_cents"`
	CurrencyCode  string `json:"currency_code"`
	CategoryID    int64  `json:"category_id"`
	CategoryName  string `json:"category_name"`
	CategoryIcon string `json:"category_icon"`
	CategoryColor string `json:"category_color"`
	Note          string `json:"note"`
	Frequency     string `json:"frequency"`
	NextRunAt     string `json:"next_run_at"`
	IsActive      bool   `json:"is_active"`
	CreatedAt     string `json:"created_at"`
}

type createRecurringRequest struct {
	AccountID    int64  `json:"account_id"`
	Type         string `json:"type"`
	AmountCents  int64  `json:"amount_cents"`
	CurrencyCode string `json:"currency_code"`
	CategoryID   int64  `json:"category_id"`
	Note         string `json:"note"`
	Frequency    string `json:"frequency"`
}

type updateRecurringRequest struct {
	AccountID    *int64 `json:"account_id"`
	Type         string `json:"type"`
	AmountCents  *int64 `json:"amount_cents"`
	CurrencyCode string `json:"currency_code"`
	CategoryID   *int64 `json:"category_id"`
	Note         string `json:"note"`
	Frequency    string `json:"frequency"`
}

// recurringHandler routes requests for /api/v1/recurring[/{id}[/toggle]].
func recurringHandler(svc *service.RecurringService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/recurring")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			parts := strings.SplitN(suffix, "/", 2)
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid recurring id"})
				return
			}

			// PATCH /api/v1/recurring/{id}/toggle
			if len(parts) == 2 && parts[1] == "toggle" {
				if r.Method != http.MethodPatch {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				rt, err := svc.ToggleActive(ctx, id, userID)
				if err != nil {
					writeError(w, log, err)
					return
				}
				writeJSON(w, http.StatusOK, recurringToResponse(rt))
				return
			}

			switch r.Method {
			case http.MethodPut:
				updateRecurring(w, r, userID, id, svc, log)
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
			listRecurring(w, r, userID, svc, log)
		case http.MethodPost:
			createRecurring(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listRecurring(w http.ResponseWriter, r *http.Request, userID int64, svc *service.RecurringService, log *slog.Logger) {
	items, err := svc.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	resp := make([]recurringResponse, 0, len(items))
	for _, rt := range items {
		resp = append(resp, recurringToResponse(rt))
	}
	writeJSON(w, http.StatusOK, map[string]any{"recurring": resp})
}

func createRecurring(w http.ResponseWriter, r *http.Request, userID int64, svc *service.RecurringService, log *slog.Logger) {
	var req createRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	rt, err := svc.Create(r.Context(), &domain.RecurringTransaction{
		UserID:       userID,
		AccountID:    req.AccountID,
		Type:         domain.TransactionType(req.Type),
		AmountCents:  req.AmountCents,
		CurrencyCode: req.CurrencyCode,
		CategoryID:   req.CategoryID,
		Note:         req.Note,
		Frequency:    domain.Frequency(req.Frequency),
	})
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, recurringToResponse(rt))
}

func updateRecurring(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.RecurringService, log *slog.Logger) {
	var req updateRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	ctx := r.Context()
	existing, err := svc.GetByID(ctx, id, userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	if req.AccountID != nil {
		existing.AccountID = *req.AccountID
	}
	if req.Type != "" {
		existing.Type = domain.TransactionType(req.Type)
	}
	if req.AmountCents != nil {
		existing.AmountCents = *req.AmountCents
	}
	if req.CurrencyCode != "" {
		existing.CurrencyCode = req.CurrencyCode
	}
	if req.CategoryID != nil {
		existing.CategoryID = *req.CategoryID
	}
	if req.Note != "" {
		existing.Note = req.Note
	}
	if req.Frequency != "" {
		existing.Frequency = domain.Frequency(req.Frequency)
	}

	rt, err := svc.Update(ctx, existing)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, recurringToResponse(rt))
}

func recurringToResponse(rt *domain.RecurringTransaction) recurringResponse {
	return recurringResponse{
		ID:            rt.ID,
		AccountID:     rt.AccountID,
		Type:          string(rt.Type),
		AmountCents:   rt.AmountCents,
		CurrencyCode:  rt.CurrencyCode,
		CategoryID:    rt.CategoryID,
		CategoryName:  rt.CategoryName,
		CategoryIcon: rt.CategoryIcon,
		CategoryColor: rt.CategoryColor,
		Note:          rt.Note,
		Frequency:     string(rt.Frequency),
		NextRunAt:     rt.NextRunAt.Format("2006-01-02T15:04:05Z"),
		IsActive:      rt.IsActive,
		CreatedAt:     rt.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
