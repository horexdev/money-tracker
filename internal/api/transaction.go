package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type transactionResponse struct {
	ID             int64  `json:"id"`
	Type           string `json:"type"`
	AmountCents    int64  `json:"amount_cents"`
	CurrencyCode   string `json:"currency_code"`
	CategoryID     int64  `json:"category_id"`
	CategoryName   string `json:"category_name"`
	CategoryEmoji  string `json:"category_emoji"`
	CategoryColor  string `json:"category_color"`
	Note           string `json:"note"`
	CreatedAt      string `json:"created_at"`
}

type listTransactionsResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	TotalPages   int                   `json:"total_pages"`
	CurrentPage  int                   `json:"current_page"`
}

type createTransactionRequest struct {
	Type         string  `json:"type"`
	AmountCents  int64   `json:"amount_cents"`
	CategoryID   int64   `json:"category_id"`
	Note         string  `json:"note"`
	CurrencyCode string  `json:"currency_code"`
	CreatedAt    *string `json:"created_at"`
}

type transactionManager interface {
	AddExpense(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode, baseCurrency string, exchangeRate float64, createdAt *time.Time) (*domain.Transaction, error)
	AddIncome(ctx context.Context, userID, amountCents, categoryID int64, note, currencyCode, baseCurrency string, exchangeRate float64, createdAt *time.Time) (*domain.Transaction, error)
	ListPaged(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Transaction, int, error)
	Delete(ctx context.Context, id, userID int64) error
	UpdateTransaction(ctx context.Context, userID, id, amountCents, categoryID int64, note string, createdAt time.Time) (*domain.Transaction, error)
}

const defaultPageSize = 20

// transactionHandler routes GET/POST/DELETE for /api/v1/transactions[/{id}]
func transactionHandler(txSvc transactionManager, userSvc *service.UserService, exchangeSvc *service.ExchangeService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		// /api/v1/transactions/{id}  → DELETE
		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/transactions")
		suffix = strings.TrimPrefix(suffix, "/")

		if suffix != "" {
			id, err := strconv.ParseInt(suffix, 10, 64)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid transaction id"})
				return
			}
			switch r.Method {
			case http.MethodDelete:
				if err := txSvc.Delete(ctx, id, userID); err != nil {
					writeError(w, log, err)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			case http.MethodPut:
				updateTransaction(w, r, userID, id, txSvc, log)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodGet:
			listTransactions(w, r, userID, txSvc, log)
		case http.MethodPost:
			createTransaction(w, r, userID, txSvc, userSvc, exchangeSvc, log)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func listTransactions(w http.ResponseWriter, r *http.Request, userID int64, txSvc transactionManager, log *slog.Logger) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = defaultPageSize
	}

	txs, totalPages, err := txSvc.ListPaged(r.Context(), userID, page, pageSize)
	if err != nil {
		writeError(w, log, err)
		return
	}

	items := make([]transactionResponse, 0, len(txs))
	for _, tx := range txs {
		items = append(items, txToResponse(tx))
	}

	writeJSON(w, http.StatusOK, listTransactionsResponse{
		Transactions: items,
		TotalPages:   totalPages,
		CurrentPage:  page,
	})
}

func createTransaction(w http.ResponseWriter, r *http.Request, userID int64, txSvc transactionManager, userSvc *service.UserService, exchangeSvc *service.ExchangeService, log *slog.Logger) {
	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}

	if req.AmountCents <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "amount_cents must be positive"})
		return
	}
	if req.CategoryID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "category_id is required"})
		return
	}

	ctx := r.Context()

	// Resolve base currency and exchange rate snapshot.
	user, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		writeError(w, log, err)
		return
	}
	baseCurrency := user.CurrencyCode
	if req.CurrencyCode == "" {
		req.CurrencyCode = baseCurrency
	}

	exchangeRate := 1.0
	if req.CurrencyCode != baseCurrency {
		rate, rateErr := exchangeSvc.GetRate(ctx, req.CurrencyCode, baseCurrency)
		if rateErr != nil {
			log.WarnContext(ctx, "exchange rate unavailable, using 1.0",
				slog.String("from", req.CurrencyCode),
				slog.String("to", baseCurrency),
				slog.String("error", rateErr.Error()),
			)
		} else {
			exchangeRate = rate
		}
	}

	var customTime *time.Time
	if req.CreatedAt != nil && *req.CreatedAt != "" {
		t, parseErr := time.Parse("2006-01-02", *req.CreatedAt)
		if parseErr != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "created_at must be in YYYY-MM-DD format"})
			return
		}
		customTime = &t
	}

	var tx *domain.Transaction
	switch req.Type {
	case "expense":
		tx, err = txSvc.AddExpense(ctx, userID, req.AmountCents, req.CategoryID, req.Note, req.CurrencyCode, baseCurrency, exchangeRate, customTime)
	case "income":
		tx, err = txSvc.AddIncome(ctx, userID, req.AmountCents, req.CategoryID, req.Note, req.CurrencyCode, baseCurrency, exchangeRate, customTime)
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "type must be 'expense' or 'income'"})
		return
	}
	if err != nil {
		writeError(w, log, err)
		return
	}

	writeJSON(w, http.StatusCreated, txToResponse(tx))
}

type updateTransactionRequest struct {
	AmountCents int64  `json:"amount_cents"`
	CategoryID  int64  `json:"category_id"`
	Note        string `json:"note"`
	CreatedAt   string `json:"created_at"`
}

func updateTransaction(w http.ResponseWriter, r *http.Request, userID, id int64, txSvc transactionManager, log *slog.Logger) {
	var req updateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
		return
	}
	if req.AmountCents <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "amount_cents must be positive"})
		return
	}
	if req.CategoryID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "category_id is required"})
		return
	}
	createdAt := time.Now()
	if req.CreatedAt != "" {
		t, err := time.Parse("2006-01-02", req.CreatedAt)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "created_at must be in YYYY-MM-DD format"})
			return
		}
		createdAt = t
	}
	tx, err := txSvc.UpdateTransaction(r.Context(), userID, id, req.AmountCents, req.CategoryID, req.Note, createdAt)
	if err != nil {
		writeError(w, log, err)
		return
	}
	writeJSON(w, http.StatusOK, txToResponse(tx))
}

func txToResponse(tx *domain.Transaction) transactionResponse {
	return transactionResponse{
		ID:            tx.ID,
		Type:          string(tx.Type),
		AmountCents:   tx.AmountCents,
		CurrencyCode:  tx.CurrencyCode,
		CategoryID:    tx.CategoryID,
		CategoryName:  tx.CategoryName,
		CategoryEmoji: tx.CategoryEmoji,
		CategoryColor: tx.CategoryColor,
		Note:          tx.Note,
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
