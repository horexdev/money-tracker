package api

import (
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

// userDataHandler handles DELETE /api/v1/user/data — resets all user data and recreates a default account.
func userDataHandler(userSvc *service.UserService, accountSvc *service.AccountService, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		userID := userIDFromContext(ctx)

		if err := userSvc.ResetData(ctx, userID); err != nil {
			writeError(w, log, err)
			return
		}

		// Recreate a default account so the user starts fresh with one account.
		user, err := userSvc.GetByID(ctx, userID)
		if err != nil {
			log.WarnContext(ctx, "userDataHandler: failed to get user after reset",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		lang := string(user.Language)
		name := localizedAccountName(lang)
		if _, err := accountSvc.Create(ctx, userID, name, "wallet", "#6366f1", domain.AccountTypeChecking, user.CurrencyCode, true); err != nil {
			log.WarnContext(ctx, "userDataHandler: failed to recreate default account",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
