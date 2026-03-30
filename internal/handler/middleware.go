package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
)

// LoggingMiddleware logs every incoming update with user ID and processing duration.
func LoggingMiddleware(log *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			start := time.Now()
			userID := extractUserID(update)

			log.InfoContext(ctx, "update received",
				slog.Int64("user_id", userID),
				slog.Int64("update_id", update.ID),
			)

			next(ctx, b, update)

			log.InfoContext(ctx, "update processed",
				slog.Int64("user_id", userID),
				slog.Int64("update_id", update.ID),
				slog.Duration("duration", time.Since(start)),
			)
		}
	}
}

// AutoRegisterMiddleware upserts the Telegram user into the DB on every update.
func AutoRegisterMiddleware(userSvc *service.UserService, log *slog.Logger) bot.Middleware {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, b *bot.Bot, update *models.Update) {
			tgUser := extractTelegramUser(update)
			if tgUser != nil {
				u := &domain.User{
					ID:        tgUser.ID,
					Username:  tgUser.Username,
					FirstName: tgUser.FirstName,
					LastName:  tgUser.LastName,
				}
				if _, err := userSvc.Upsert(ctx, u); err != nil {
					log.ErrorContext(ctx, "auto-register failed",
						slog.Int64("user_id", tgUser.ID),
						slog.String("error", err.Error()),
					)
				}
			}
			next(ctx, b, update)
		}
	}
}

// extractUserID pulls the user ID from any supported update type.
func extractUserID(update *models.Update) int64 {
	if u := extractTelegramUser(update); u != nil {
		return u.ID
	}
	return 0
}

func extractTelegramUser(update *models.Update) *models.User {
	switch {
	case update.Message != nil && update.Message.From != nil:
		return update.Message.From
	case update.CallbackQuery != nil && update.CallbackQuery.From.ID != 0:
		return &update.CallbackQuery.From
	default:
		return nil
	}
}
