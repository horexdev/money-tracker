package handler

import (
	"context"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// sendWithMainMenu is a convenience wrapper that sends an HTML message with the main menu keyboard.
func sendWithMainMenu(ctx context.Context, b *bot.Bot, chatID int64, text string, log *slog.Logger) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: mainMenuKeyboard(),
	}); err != nil {
		log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
	}
}

// sendWithCancel sends an HTML message with the cancel-only reply keyboard.
func sendWithCancel(ctx context.Context, b *bot.Bot, chatID int64, text string, log *slog.Logger) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: cancelKeyboard(),
	}); err != nil {
		log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
	}
}

// sendWithInline sends an HTML message with an inline keyboard and cancel reply keyboard.
func sendWithInline(ctx context.Context, b *bot.Bot, chatID int64, text string, inline *models.InlineKeyboardMarkup, log *slog.Logger) {
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: inline,
	}); err != nil {
		log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
	}
}
