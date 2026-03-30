package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/service"
)

// RegisterAll attaches the /start and /help handlers to the bot.
func RegisterAll(b *bot.Bot, miniAppURL string, userSvc *service.UserService, log *slog.Logger) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, StartHandler(miniAppURL, userSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, HelpHandler(miniAppURL, userSvc, log))
}

// DefaultHandler is the fallback for any unrecognised message.
func DefaultHandler(miniAppURL string, userSvc *service.UserService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		lang := userLang(ctx, userSvc, update.Message.From.ID)
		s := getString(lang)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   s.defaultMsg,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{Text: s.openButton, WebApp: &models.WebAppInfo{URL: miniAppURL}}},
				},
			},
		}); err != nil {
			log.ErrorContext(ctx, "default: failed to send message", slog.String("error", err.Error()))
		}
	}
}
