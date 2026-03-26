package handler

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const welcomeText = `👋 Welcome to *MoneyTracker*!

Track your income and expenses right here in Telegram.

*Commands:*
/addexpense — record an expense
/addincome  — record income
/balance    — show your current balance
/history    — show last 10 transactions
/stats      — view stats by category
/cancel     — cancel current action
/help       — show this message`

// StartHandler handles the /start and /help commands.
func StartHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      welcomeText,
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
