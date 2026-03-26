package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/repository"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

// IncomeStartHandler initiates the add-income flow.
func IncomeStartHandler(store *fsm.Store, catRepo *repository.CategoryRepository, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		if err := store.SetState(ctx, userID, fsm.StateIncomeWaitAmount); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "💰 Enter income amount (e.g. 5000.00):",
		}); err != nil {
			log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		}
	}
}

// IncomeAmountHandler processes the amount step of the income flow.
func IncomeAmountHandler(store *fsm.Store, catRepo *repository.CategoryRepository, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		text := update.Message.Text

		cents, err := money.ParseCents(text)
		if err != nil {
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Invalid amount. Please enter a positive number like 5000.00:",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
			return
		}

		if err := store.SetData(ctx, userID, "amount", strconv.FormatInt(cents, 10)); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		if err := store.SetState(ctx, userID, fsm.StateIncomeWaitCategory); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		cats, err := catRepo.ListForUser(ctx, userID)
		if err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "📂 Choose a category:",
			ReplyMarkup: buildCategoryKeyboard(cats),
		}); err != nil {
			log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		}
	}
}

// IncomeCategoryHandler processes the category callback for the income flow.
func IncomeCategoryHandler(store *fsm.Store, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID
		categoryID := parseCategoryCallback(query.Data)

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		if err := store.SetData(ctx, userID, "category", strconv.FormatInt(categoryID, 10)); err != nil {
			sendErrorCallback(ctx, b, query.Message.Message.Chat.ID)
			return
		}
		if err := store.SetState(ctx, userID, fsm.StateIncomeWaitNote); err != nil {
			sendErrorCallback(ctx, b, query.Message.Message.Chat.ID)
			return
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: query.Message.Message.Chat.ID,
			Text:   "📝 Add a note (or send /skip):",
		}); err != nil {
			log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		}
	}
}

// IncomeNoteHandler finalises the income transaction.
func IncomeNoteHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		note := update.Message.Text
		if note == "/skip" {
			note = ""
		}

		amountStr, _ := store.GetData(ctx, userID, "amount")
		categoryStr, _ := store.GetData(ctx, userID, "category")

		cents, _ := strconv.ParseInt(amountStr, 10, 64)
		catID, _ := strconv.ParseInt(categoryStr, 10, 64)

		tx, err := txSvc.AddIncome(ctx, userID, cents, catID, note)
		if err != nil {
			log.ErrorContext(ctx, "add income failed", slog.String("error", err.Error()))
			sendError(ctx, b, update.Message.Chat.ID)
		} else {
			emoji := tx.CategoryEmoji
			if emoji == "" {
				emoji = "📦"
			}
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: fmt.Sprintf("✅ Income recorded: +%s %s %s",
					money.FormatCents(tx.AmountCents), emoji, tx.CategoryName),
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
		}

		if err := store.Clear(ctx, userID); err != nil {
			log.ErrorContext(ctx, "failed to clear FSM state", slog.String("error", err.Error()))
		}
	}
}
