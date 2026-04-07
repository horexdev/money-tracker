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

type accountResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	Type           string `json:"type"`
	CurrencyCode   string `json:"currency_code"`
	IsDefault      bool   `json:"is_default"`
	IncludeInTotal bool   `json:"include_in_total"`
	BalanceCents   int64  `json:"balance_cents"`
	CreatedAt      string `json:"created_at"`
}

type createAccountRequest struct {
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	Type           string `json:"type"`
	CurrencyCode   string `json:"currency_code"`
	IncludeInTotal *bool  `json:"include_in_total"`
}

type updateAccountRequest struct {
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	Type           string `json:"type"`
	CurrencyCode   string `json:"currency_code"`
	IncludeInTotal *bool  `json:"include_in_total"`
}

// accountsHandler routes requests for /api/v1/accounts[/{id}[/set-default|/adjust]].
func accountsHandler(svc *service.AccountService, adjustSvc *service.AdjustmentService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/accounts")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			parts := strings.SplitN(suffix, "/", 2)
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid account id"})
				return
			}

			if len(parts) == 2 {
				switch parts[1] {
				case "set-default":
					if r.Method != http.MethodPost {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					a, err := svc.SetDefault(ctx, id, userID)
					if err != nil {
						writeError(w, log, err)
						return
					}
					writeJSON(w, http.StatusOK, accountToResponse(a))
				case "adjust":
					adjustAccountHandler(adjustSvc, log, id)(w, r)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
				return
			}

			switch r.Method {
			case http.MethodGet:
				a, err := svc.GetByID(ctx, id, userID)
				if err != nil {
					writeError(w, log, err)
					return
				}
				writeJSON(w, http.StatusOK, accountToResponse(a))
			case http.MethodPut:
				updateAccount(w, r, userID, id, svc, log)
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
			listAccounts(w, r, userID, svc, log)
		case http.MethodPost:
			createAccount(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listAccounts(w http.ResponseWriter, r *http.Request, userID int64, svc *service.AccountService, log *slog.Logger) {
	accounts, err := svc.List(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}
	items := make([]accountResponse, 0, len(accounts))
	for _, a := range accounts {
		items = append(items, accountToResponse(a))
	}
	writeJSON(w, http.StatusOK, map[string]any{"accounts": items})
}

func createAccount(w http.ResponseWriter, r *http.Request, userID int64, svc *service.AccountService, log *slog.Logger) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "name is required"})
		return
	}
	if req.Icon == "" {
		req.Icon = "wallet"
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}
	if req.Type == "" {
		req.Type = "checking"
	}
	if req.CurrencyCode == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "currency_code is required"})
		return
	}
	includeInTotal := true
	if req.IncludeInTotal != nil {
		includeInTotal = *req.IncludeInTotal
	}

	a, err := svc.Create(r.Context(), userID, req.Name, req.Icon, req.Color,
		domain.AccountType(req.Type), req.CurrencyCode, includeInTotal)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, accountToResponse(a))
}

func updateAccount(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.AccountService, log *slog.Logger) {
	var req updateAccountRequest
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

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Icon != "" {
		existing.Icon = req.Icon
	}
	if req.Color != "" {
		existing.Color = req.Color
	}
	if req.Type != "" {
		existing.Type = domain.AccountType(req.Type)
	}
	if req.CurrencyCode != "" {
		existing.CurrencyCode = req.CurrencyCode
	}
	if req.IncludeInTotal != nil {
		existing.IncludeInTotal = *req.IncludeInTotal
	}

	updated, err := svc.Update(ctx, existing.ID, userID, existing.Name, existing.Icon,
		existing.Color, existing.Type, existing.CurrencyCode, existing.IncludeInTotal)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, accountToResponse(updated))
}

func accountToResponse(a *domain.Account) accountResponse {
	return accountResponse{
		ID:             a.ID,
		Name:           a.Name,
		Icon:           a.Icon,
		Color:          a.Color,
		Type:           string(a.Type),
		CurrencyCode:   a.CurrencyCode,
		IsDefault:      a.IsDefault,
		IncludeInTotal: a.IncludeInTotal,
		BalanceCents:   a.BalanceCents,
		CreatedAt:      a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
