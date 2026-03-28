package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// RegisterAll attaches the /start and /help handlers to the bot.
func RegisterAll(b *bot.Bot, miniAppURL string, log *slog.Logger) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, StartHandler(miniAppURL, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, HelpHandler(miniAppURL, log))
}

// DefaultHandler is the fallback for any unrecognized message.
// It directs the user to open the Mini App.
func DefaultHandler(miniAppURL string, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Use the button below to open MoneyTracker, or type /help for more info.",
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
			log.ErrorContext(ctx, "default: failed to send message", slog.String("error", err.Error()))
		}
	}
}
