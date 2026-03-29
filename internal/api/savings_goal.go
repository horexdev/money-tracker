package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type goalResponse struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	TargetCents     int64   `json:"target_cents"`
	CurrentCents    int64   `json:"current_cents"`
	CurrencyCode    string  `json:"currency_code"`
	Deadline        *string `json:"deadline"`
	ProgressPercent float64 `json:"progress_percent"`
	IsCompleted     bool    `json:"is_completed"`
	RemainingCents  int64   `json:"remaining_cents"`
	CreatedAt       string  `json:"created_at"`
}

type createGoalRequest struct {
	Name         string `json:"name"`
	TargetCents  int64  `json:"target_cents"`
	CurrencyCode string `json:"currency_code"`
	Deadline     string `json:"deadline"`
}

type updateGoalRequest struct {
	Name        string `json:"name"`
	TargetCents *int64 `json:"target_cents"`
	Deadline    string `json:"deadline"`
}

type depositWithdrawRequest struct {
	AmountCents int64 `json:"amount_cents"`
}

// goalsHandler routes requests for /api/v1/goals[/{id}[/deposit|/withdraw]].
func goalsHandler(svc *service.SavingsGoalService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/goals")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			parts := strings.SplitN(suffix, "/", 2)
			id, err := strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid goal id"})
				return
			}

			if len(parts) == 2 {
				switch parts[1] {
				case "deposit":
					if r.Method != http.MethodPost {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					goalDeposit(w, r, userID, id, svc, log)
				case "withdraw":
					if r.Method != http.MethodPost {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					goalWithdraw(w, r, userID, id, svc, log)
				case "history":
					if r.Method != http.MethodGet {
						w.WriteHeader(http.StatusMethodNotAllowed)
						return
					}
					goalHistory(w, r, userID, id, svc, log)
				default:
					w.WriteHeader(http.StatusNotFound)
				}
				return
			}

			switch r.Method {
			case http.MethodPut:
				updateGoal(w, r, userID, id, svc, log)
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
			listGoals(w, r, userID, svc, log)
		case http.MethodPost:
			createGoal(w, r, userID, svc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listGoals(w http.ResponseWriter, r *http.Request, userID int64, svc *service.SavingsGoalService, log *slog.Logger) {
	goals, err := svc.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]goalResponse, 0, len(goals))
	for _, g := range goals {
		items = append(items, goalToResponse(g))
	}
	writeJSON(w, http.StatusOK, map[string]any{"goals": items})
}

func createGoal(w http.ResponseWriter, r *http.Request, userID int64, svc *service.SavingsGoalService, log *slog.Logger) {
	var req createGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	goal := &domain.SavingsGoal{
		UserID:       userID,
		Name:         req.Name,
		TargetCents:  req.TargetCents,
		CurrencyCode: req.CurrencyCode,
	}

	if req.Deadline != "" {
		t, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid deadline format, use YYYY-MM-DD"})
			return
		}
		goal.Deadline = &t
	}

	result, err := svc.Create(r.Context(), goal)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusCreated, goalToResponse(result))
}

func updateGoal(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.SavingsGoalService, log *slog.Logger) {
	var req updateGoalRequest
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
	if req.TargetCents != nil {
		existing.TargetCents = *req.TargetCents
	}
	if req.Deadline != "" {
		t, err := time.Parse("2006-01-02", req.Deadline)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid deadline format, use YYYY-MM-DD"})
			return
		}
		existing.Deadline = &t
	}

	result, err := svc.Update(ctx, existing)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, goalToResponse(result))
}

func goalDeposit(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.SavingsGoalService, log *slog.Logger) {
	var req depositWithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	goal, err := svc.Deposit(r.Context(), id, userID, req.AmountCents)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, goalToResponse(goal))
}

func goalWithdraw(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.SavingsGoalService, log *slog.Logger) {
	var req depositWithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	goal, err := svc.Withdraw(r.Context(), id, userID, req.AmountCents)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, goalToResponse(goal))
}

type goalTransactionResponse struct {
	ID          int64  `json:"id"`
	Type        string `json:"type"`
	AmountCents int64  `json:"amount_cents"`
	CreatedAt   string `json:"created_at"`
}

func goalHistory(w http.ResponseWriter, r *http.Request, userID, id int64, svc *service.SavingsGoalService, log *slog.Logger) {
	history, err := svc.ListHistory(r.Context(), id, userID)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]goalTransactionResponse, 0, len(history))
	for _, tx := range history {
		items = append(items, goalTransactionResponse{
			ID:          tx.ID,
			Type:        tx.Type,
			AmountCents: tx.AmountCents,
			CreatedAt:   tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": items})
}

func goalToResponse(g *domain.SavingsGoal) goalResponse {
	var deadline *string
	if g.Deadline != nil {
		s := g.Deadline.Format("2006-01-02")
		deadline = &s
	}
	return goalResponse{
		ID:              g.ID,
		Name:            g.Name,
		TargetCents:     g.TargetCents,
		CurrentCents:    g.CurrentCents,
		CurrencyCode:    g.CurrencyCode,
		Deadline:        deadline,
		ProgressPercent: g.ProgressPercent(),
		IsCompleted:     g.IsCompleted(),
		RemainingCents:  g.RemainingCents(),
		CreatedAt:       g.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
