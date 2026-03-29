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

type budgetResponse struct {
	ID              int64   `json:"id"`
	CategoryID      int64   `json:"category_id"`
	CategoryName    string  `json:"category_name"`
	CategoryEmoji   string  `json:"category_emoji"`
	CategoryColor   string  `json:"category_color"`
	LimitCents      int64   `json:"limit_cents"`
	SpentCents      int64   `json:"spent_cents"`
	Period          string  `json:"period"`
	CurrencyCode    string  `json:"currency_code"`
	NotifyAtPercent int     `json:"notify_at_percent"`
	UsagePercent    float64 `json:"usage_percent"`
	IsOverLimit     bool    `json:"is_over_limit"`
}

type createBudgetRequest struct {
	CategoryID      int64  `json:"category_id"`
	LimitCents      int64  `json:"limit_cents"`
	Period          string `json:"period"`
	CurrencyCode    string `json:"currency_code"`
	NotifyAtPercent *int   `json:"notify_at_percent"`
}

type updateBudgetRequest struct {
	LimitCents      *int64 `json:"limit_cents"`
	Period          string `json:"period"`
	NotifyAtPercent *int   `json:"notify_at_percent"`
}

// budgetHandler routes requests for /api/v1/budgets[/{id}[/transactions]].
func budgetHandler(svc *service.BudgetService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/budgets")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			parts := strings.SplitN(suffix, "/", 2)
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid budget id"})
				return
			}

			if len(parts) == 2 {
				switch parts[1] {
				case "transactions":
					if r.Method != http.MethodGet {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					listBudgetTransactions(w, r, userID, id, svc, log)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
				return
			}

			switch r.Method {
			case http.MethodPut:
				updateBudget(w, r, userID, id, svc, log)
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
			listBudgets(w, r, userID, svc, log)
		case http.MethodPost:
			createBudget(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listBudgets(w http.ResponseWriter, r *http.Request, userID int64, svc *service.BudgetService, log *slog.Logger) {
	budgets, err := svc.ListWithProgress(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]budgetResponse, 0, len(budgets))
	for _, b := range budgets {
		items = append(items, budgetToResponse(b))
	}
	writeJSON(w, http.StatusOK, map[string]any{"budgets": items})
}

func createBudget(w http.ResponseWriter, r *http.Request, userID int64, svc *service.BudgetService, log *slog.Logger) {
	var req createBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	notifyAt := 80
	if req.NotifyAtPercent != nil {
		notifyAt = *req.NotifyAtPercent
	}

	budget, err := svc.Create(r.Context(), &domain.Budget{
		UserID:          userID,
		CategoryID:      req.CategoryID,
		LimitCents:      req.LimitCents,
		Period:          domain.BudgetPeriod(req.Period),
		CurrencyCode:    req.CurrencyCode,
		NotifyAtPercent: notifyAt,
	})
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, budgetToResponse(budget))
}

func updateBudget(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.BudgetService, log *slog.Logger) {
	var req updateBudgetRequest
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

	if req.LimitCents != nil {
		existing.LimitCents = *req.LimitCents
	}
	if req.Period != "" {
		existing.Period = domain.BudgetPeriod(req.Period)
	}
	if req.NotifyAtPercent != nil {
		existing.NotifyAtPercent = *req.NotifyAtPercent
	}

	budget, err := svc.Update(ctx, existing)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, budgetToResponse(budget))
}

func listBudgetTransactions(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.BudgetService, log *slog.Logger) {
	txs, err := svc.ListForBudget(r.Context(), id, userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]budgetTransactionResponse, 0, len(txs))
	for _, tx := range txs {
		items = append(items, budgetTransactionResponse{
			ID:            tx.ID,
			AmountCents:   tx.AmountCents,
			CategoryName:  tx.CategoryName,
			CategoryEmoji: tx.CategoryEmoji,
			CategoryColor: tx.CategoryColor,
			Note:          tx.Note,
			CurrencyCode:  tx.CurrencyCode,
			CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"transactions": items})
}

type budgetTransactionResponse struct {
	ID            int64  `json:"id"`
	AmountCents   int64  `json:"amount_cents"`
	CategoryName  string `json:"category_name"`
	CategoryEmoji string `json:"category_emoji"`
	CategoryColor string `json:"category_color"`
	Note          string `json:"note"`
	CurrencyCode  string `json:"currency_code"`
	CreatedAt     string `json:"created_at"`
}

func budgetToResponse(b *domain.Budget) budgetResponse {
	return budgetResponse{
		ID:              b.ID,
		CategoryID:      b.CategoryID,
		CategoryName:    b.CategoryName,
		CategoryEmoji:   b.CategoryEmoji,
		CategoryColor:   b.CategoryColor,
		LimitCents:      b.LimitCents,
		SpentCents:      b.SpentCents,
		Period:          string(b.Period),
		CurrencyCode:    b.CurrencyCode,
		NotifyAtPercent: b.NotifyAtPercent,
		UsagePercent:    b.UsagePercent(),
		IsOverLimit:     b.IsOverLimit(),
	}
}
