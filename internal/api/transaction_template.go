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

type templateResponse struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	AmountCents   int64  `json:"amount_cents"`
	AmountFixed   bool   `json:"amount_fixed"`
	CurrencyCode  string `json:"currency_code"`
	CategoryID    int64  `json:"category_id"`
	CategoryName  string `json:"category_name"`
	CategoryIcon  string `json:"category_icon"`
	CategoryColor string `json:"category_color"`
	AccountID     int64  `json:"account_id"`
	Note          string `json:"note"`
	SortOrder     int32  `json:"sort_order"`
	CreatedAt     string `json:"created_at"`
}

type createTemplateRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	AmountCents  int64  `json:"amount_cents"`
	AmountFixed  bool   `json:"amount_fixed"`
	CurrencyCode string `json:"currency_code"`
	CategoryID   int64  `json:"category_id"`
	AccountID    int64  `json:"account_id"`
	Note         string `json:"note"`
}

type updateTemplateRequest struct {
	Name         *string `json:"name"`
	Type         string  `json:"type"`
	AmountCents  *int64  `json:"amount_cents"`
	AmountFixed  *bool   `json:"amount_fixed"`
	CurrencyCode string  `json:"currency_code"`
	CategoryID   *int64  `json:"category_id"`
	AccountID    *int64  `json:"account_id"`
	Note         *string `json:"note"`
}

type applyTemplateRequest struct {
	AmountCents *int64 `json:"amount_cents"`
}

type reorderTemplatesRequest struct {
	Order []int64 `json:"order"`
}

// templateHandler routes requests for /api/v1/templates[/{id}[/apply]] and /api/v1/templates/reorder.
func templateHandler(svc *service.TransactionTemplateService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/templates")
		suffix = strings.TrimPrefix(suffix, "/")

		// PATCH /api/v1/templates/reorder
		if suffix == "reorder" {
			if r.Method != http.MethodPatch {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			reorderTemplates(w, r, userID, svc, log)
			return
		}

		if suffix != "" {
			parts := strings.SplitN(suffix, "/", 2)
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid template id"})
				return
			}

			// POST /api/v1/templates/{id}/apply
			if len(parts) == 2 && parts[1] == "apply" {
				if r.Method != http.MethodPost {
					w.WriteHeader(http.StatusMethodNotAllowed)
					return
				}
				applyTemplate(w, r, userID, id, svc, log)
				return
			}

			switch r.Method {
			case http.MethodPut:
				updateTemplate(w, r, userID, id, svc, log)
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
			listTemplates(w, r, userID, svc, log)
		case http.MethodPost:
			createTemplate(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listTemplates(w http.ResponseWriter, r *http.Request, userID int64, svc *service.TransactionTemplateService, log *slog.Logger) {
	items, err := svc.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}
	resp := make([]templateResponse, 0, len(items))
	for _, t := range items {
		resp = append(resp, templateToResponse(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": resp})
}

func createTemplate(w http.ResponseWriter, r *http.Request, userID int64, svc *service.TransactionTemplateService, log *slog.Logger) {
	var req createTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	t, err := svc.Create(r.Context(), &domain.TransactionTemplate{
		UserID:       userID,
		Name:         req.Name,
		Type:         domain.TransactionType(req.Type),
		AmountCents:  req.AmountCents,
		AmountFixed:  req.AmountFixed,
		CurrencyCode: req.CurrencyCode,
		CategoryID:   req.CategoryID,
		AccountID:    req.AccountID,
		Note:         req.Note,
	})
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, templateToResponse(t))
}

func updateTemplate(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.TransactionTemplateService, log *slog.Logger) {
	var req updateTemplateRequest
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
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Type != "" {
		existing.Type = domain.TransactionType(req.Type)
	}
	if req.AmountCents != nil {
		existing.AmountCents = *req.AmountCents
	}
	if req.AmountFixed != nil {
		existing.AmountFixed = *req.AmountFixed
	}
	if req.CurrencyCode != "" {
		existing.CurrencyCode = req.CurrencyCode
	}
	if req.CategoryID != nil {
		existing.CategoryID = *req.CategoryID
	}
	if req.AccountID != nil {
		existing.AccountID = *req.AccountID
	}
	if req.Note != nil {
		existing.Note = *req.Note
	}
	updated, err := svc.Update(ctx, existing)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, templateToResponse(updated))
}

func applyTemplate(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.TransactionTemplateService, log *slog.Logger) {
	var req applyTemplateRequest
	// Empty body is valid (fixed-amount template).
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
			return
		}
	}
	tx, err := svc.Apply(r.Context(), id, userID, req.AmountCents)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, txToResponse(tx))
}

func reorderTemplates(w http.ResponseWriter, r *http.Request, userID int64, svc *service.TransactionTemplateService, log *slog.Logger) {
	var req reorderTemplatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if err := svc.Reorder(r.Context(), userID, req.Order); err != nil {
		writeError(w, log, err)
		return
	}
	items, err := svc.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}
	resp := make([]templateResponse, 0, len(items))
	for _, t := range items {
		resp = append(resp, templateToResponse(t))
	}
	writeJSON(w, http.StatusOK, map[string]any{"templates": resp})
}

func templateToResponse(t *domain.TransactionTemplate) templateResponse {
	return templateResponse{
		ID:            t.ID,
		Name:          t.Name,
		Type:          string(t.Type),
		AmountCents:   t.AmountCents,
		AmountFixed:   t.AmountFixed,
		CurrencyCode:  t.CurrencyCode,
		CategoryID:    t.CategoryID,
		CategoryName:  t.CategoryName,
		CategoryIcon:  t.CategoryIcon,
		CategoryColor: t.CategoryColor,
		AccountID:     t.AccountID,
		Note:          t.Note,
		SortOrder:     t.SortOrder,
		CreatedAt:     t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
