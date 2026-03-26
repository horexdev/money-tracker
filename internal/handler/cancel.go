package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/fsm"
)

// CancelHandler aborts any active FSM flow for the user.
func CancelHandler(store *fsm.Store, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID

		if err := store.Clear(ctx, userID); err != nil {
			log.ErrorContext(ctx, "failed to clear FSM state on cancel",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
		}

		sendWithMainMenu(ctx, b, update.Message.Chat.ID, "Cancelled. What would you like to do?", log)
	}
}
