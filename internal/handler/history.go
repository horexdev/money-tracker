package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

const historyLimit = 10

// HistoryHandler shows the user's last 10 transactions.
func HistoryHandler(txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID

		txs, err := txSvc.ListRecent(ctx, userID, historyLimit)
		if err != nil {
			log.ErrorContext(ctx, "failed to list transactions",
				slog.Int64("user_id", userID),
				slog.String("error", err.Error()),
			)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		if len(txs) == 0 {
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "No transactions yet. Use /addexpense or /addincome to get started.",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
			return
		}

		var sb strings.Builder
		sb.WriteString("📋 *Last transactions:*\n\n")

		for _, tx := range txs {
			sign := "-"
			if tx.Type == domain.TransactionTypeIncome {
				sign = "+"
			}
			emoji := tx.CategoryEmoji
			if emoji == "" {
				emoji = "📦"
			}
			note := ""
			if tx.Note != "" {
				note = " — " + tx.Note
			}
			fmt.Fprintf(&sb, "%s %s `%s%s`%s\n",
				emoji, tx.CategoryName, sign, money.FormatCents(tx.AmountCents), note,
			)
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      sb.String(),
			ParseMode: models.ParseModeMarkdown,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send history", slog.String("error", err.Error()))
		}
	}
}
