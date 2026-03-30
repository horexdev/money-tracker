package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

// userLang fetches the stored language for the given user, falling back to English on any error.
func userLang(ctx context.Context, userSvc *service.UserService, userID int64) domain.Language {
	u, err := userSvc.GetByID(ctx, userID)
	if err != nil || u == nil {
		return domain.LangEN
	}
	if u.Language == "" {
		return domain.LangEN
	}
	return u.Language
}

// StartHandler handles the /start command — sends a localised welcome message with a Mini App button.
func StartHandler(miniAppURL string, userSvc *service.UserService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		lang := userLang(ctx, userSvc, update.Message.From.ID)
		s := getString(lang)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      s.welcome,
			ParseMode: models.ParseModeHTML,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{Text: s.openButton, WebApp: &models.WebAppInfo{URL: miniAppURL}}},
				},
			},
		}); err != nil {
			log.ErrorContext(ctx, "start: failed to send message", slog.String("error", err.Error()))
		}
	}
}

// HelpHandler handles the /help command — sends a localised help message with a Mini App button.
func HelpHandler(miniAppURL string, userSvc *service.UserService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		lang := userLang(ctx, userSvc, update.Message.From.ID)
		s := getString(lang)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      s.help,
			ParseMode: models.ParseModeHTML,
			ReplyMarkup: &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{Text: s.openButton, WebApp: &models.WebAppInfo{URL: miniAppURL}}},
				},
			},
		}); err != nil {
			log.ErrorContext(ctx, "help: failed to send message", slog.String("error", err.Error()))
		}
	}
}
