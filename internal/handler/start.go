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

Tap the button below to open the app 👇`

const helpText = `<b>MoneyTracker Help</b>

Open the Mini App to manage your finances:
• Add expenses and income
• View balance and transaction history
• Track budgets and savings goals
• Set up recurring transactions
• Export your data

Use the button below to get started.`

// StartHandler handles the /start command — sends a welcome message with a Mini App button.
func StartHandler(miniAppURL string, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      welcomeText,
			ParseMode: models.ParseModeHTML,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{
							Text:   "📱 Open MoneyTracker",
							WebApp: &models.WebAppInfo{URL: miniAppURL},
						},
					},
				},
			},
		}); err != nil {
			log.ErrorContext(ctx, "start: failed to send message", slog.String("error", err.Error()))
		}
	}
}

// HelpHandler handles the /help command — sends help text with a Mini App button.
func HelpHandler(miniAppURL string, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      helpText,
			ParseMode: models.ParseModeHTML,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{
						{
							Text:   "📱 Open MoneyTracker",
							WebApp: &models.WebAppInfo{URL: miniAppURL},
						},
					},
				},
			},
		}); err != nil {
			log.ErrorContext(ctx, "help: failed to send message", slog.String("error", err.Error()))
		}
	}
}
