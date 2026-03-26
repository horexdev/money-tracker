package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

// BalanceHandler shows the user's current net balance.
func BalanceHandler(txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID

		income, expense, err := txSvc.GetBalance(ctx, userID)
		if err != nil {
			log.ErrorContext(ctx, "failed to get balance",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		net := income - expense
		text := fmt.Sprintf(
			"💰 *Balance*\n\n"+
				"Income:  `+%s`\n"+
				"Expense: `-%s`\n"+
				"───────────\n"+
				"Net:     `%s`",
			money.FormatCents(income),
			money.FormatCents(expense),
			formatNet(net),
		)

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      text,
			ParseMode: models.ParseModeMarkdown,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send balance", slog.String("error", err.Error()))
		}
	}
}

func formatNet(cents int64) string {
	if cents >= 0 {
		return "+" + money.FormatCents(cents)
	}
	return money.FormatCents(cents)
}
