package api

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

type balanceCurrencyResponse struct {
	CurrencyCode string `json:"currency_code"`
	IncomeCents  int64  `json:"income_cents"`
	ExpenseCents int64  `json:"expense_cents"`
	NetCents     int64  `json:"net_cents"`
}

type displayConversion struct {
	CurrencyCode string `json:"currency_code"`
	NetCents     int64  `json:"net_cents"`
}

type balanceResponse struct {
	ByCurrency        []balanceCurrencyResponse `json:"by_currency"`
	DisplayConversions []displayConversion       `json:"display_conversions"`
	TotalInBaseCents  int64                     `json:"total_in_base_cents"`
}

type balanceFetcher interface {
	GetBalanceByCurrency(ctx context.Context, userID int64) ([]domain.BalanceByCurrency, error)
	GetBalanceByCurrencyAndAccount(ctx context.Context, userID, accountID int64) ([]domain.BalanceByCurrency, error)
	GetTotalInBaseCurrency(ctx context.Context, userID int64) (int64, error)
}

func balanceHandler(txSvc balanceFetcher, userSvc *service.UserService, exchangeSvc *service.ExchangeService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx := r.Context()
		userID := userIDFromContext(ctx)

		var balances []domain.BalanceByCurrency
		var err error
		if accountIDStr := r.URL.Query().Get("account_id"); accountIDStr != "" {
			if accountID, parseErr := strconv.ParseInt(accountIDStr, 10, 64); parseErr == nil && accountID > 0 {
				balances, err = txSvc.GetBalanceByCurrencyAndAccount(ctx, userID, accountID)
			} else {
				balances, err = txSvc.GetBalanceByCurrency(ctx, userID)
			}
		} else {
			balances, err = txSvc.GetBalanceByCurrency(ctx, userID)
		}
		if err != nil {
			writeError(w, log, err)
			return
		}

		user, err := userSvc.GetByID(ctx, userID)
		if err != nil {
			writeError(w, log, err)
			return
		}

		byCurrency := make([]balanceCurrencyResponse, 0, len(balances))
		var baseNetCents int64
		for _, b := range balances {
			net := b.IncomeCents - b.ExpenseCents
			byCurrency = append(byCurrency, balanceCurrencyResponse{
				CurrencyCode: b.CurrencyCode,
				IncomeCents:  b.IncomeCents,
				ExpenseCents: b.ExpenseCents,
				NetCents:     net,
			})
			if b.CurrencyCode == user.CurrencyCode {
				baseNetCents = net
			}
		}

		var displayConversions []displayConversion
		if len(user.DisplayCurrencies) > 0 {
			converted, err := exchangeSvc.ConvertMulti(ctx, baseNetCents, user.CurrencyCode, user.DisplayCurrencies)
			if err != nil {
				log.WarnContext(ctx, "balance: exchange conversion failed", slog.String("error", err.Error()))
			} else {
				for _, code := range user.DisplayCurrencies {
					if netCents, ok := converted[code]; ok {
						displayConversions = append(displayConversions, displayConversion{
							CurrencyCode: code,
							NetCents:     netCents,
						})
					}
				}
			}
		}

		totalInBase, err := txSvc.GetTotalInBaseCurrency(ctx, userID)
		if err != nil {
			log.WarnContext(ctx, "balance: get total in base currency failed", slog.String("error", err.Error()))
		}

		writeJSON(w, http.StatusOK, balanceResponse{
			ByCurrency:         byCurrency,
			DisplayConversions: displayConversions,
			TotalInBaseCents:   totalInBase,
		})
	}
}
