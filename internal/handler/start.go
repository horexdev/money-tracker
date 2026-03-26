package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const welcomeText = `<b>Welcome to MoneyTracker!</b>

Your personal finance companion in Telegram.
Track expenses, monitor income, and stay on top of your money.

<b>Quick guide:</b>
💸 <b>Expense</b> — record a purchase
💰 <b>Income</b> — record earnings
💳 <b>Balance</b> — see your summary
📋 <b>History</b> — browse transactions
📊 <b>Stats</b> — spending breakdown
⚙️ <b>Settings</b> — preferences
📓 <b>Devblog</b> — release notes

Tap any button below to get started 👇`

// StartHandler handles the /start and /help commands.
func StartHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        welcomeText,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: mainMenuKeyboard(),
		}); err != nil {
			log.ErrorContext(ctx, "start: failed to send message", slog.String("error", err.Error()))
		}
	}
}
