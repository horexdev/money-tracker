package api

import (
	"log/slog"
	"net/http"

	"github.com/horexdev/money-tracker/internal/service"
)

// userDataHandler handles DELETE /api/v1/user/data — resets all user data.
func userDataHandler(userSvc *service.UserService, log *slog.Logger) http.HandlerFunc {
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

		w.WriteHeader(http.StatusNoContent)
	}
}
