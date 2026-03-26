package handler

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/service"
)

// NavHandler dispatches navigation callbacks from inline action buttons.
// Supported routes: "nav:expense", "nav:income", "nav:stats", "nav:balance".
func NavHandler(
	store *fsm.Store,
	txSvc *service.TransactionService,
	userSvc *service.UserService,
	statsSvc *service.StatsService,
	exchangeSvc *service.ExchangeService,
	log *slog.Logger,
) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		chatID := query.Message.Message.Chat.ID
		route := strings.TrimPrefix(query.Data, "nav:")

		switch route {
		case "expense":
			startTransactionFlow(ctx, b, store, userSvc, query.From.ID, chatID, "expense", log)
		case "income":
			startTransactionFlow(ctx, b, store, userSvc, query.From.ID, chatID, "income", log)
		case "stats":
			sendStatsMenu(ctx, b, chatID, log)
		case "balance":
			sendBalance(ctx, b, txSvc, userSvc, exchangeSvc, query.From.ID, chatID, log)
		}
	}
}

// NoopHandler handles non-interactive buttons (e.g. page counter).
func NoopHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
		}); err != nil {
			log.ErrorContext(ctx, "failed to answer noop callback", slog.String("error", err.Error()))
		}
	}
}
