package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

// ExpenseStartHandler initiates the add-expense flow.
func ExpenseStartHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		if err := store.SetState(ctx, userID, fsm.StateExpenseWaitAmount); err != nil {
			log.ErrorContext(ctx, "set state failed", slog.String("error", err.Error()))
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "💸 Enter expense amount (e.g. 12.50):",
		}); err != nil {
			log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		}
	}
}

// ExpenseAmountHandler processes the amount step of the expense flow.
func ExpenseAmountHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		text := update.Message.Text

		cents, err := money.ParseCents(text)
		if err != nil {
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "❌ Invalid amount. Please enter a positive number like 12.50:",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
			return
		}

		if err := store.SetData(ctx, userID, "amount", strconv.FormatInt(cents, 10)); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		if err := store.SetState(ctx, userID, fsm.StateExpenseWaitCategory); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		cats, err := txSvc.ListCategories(ctx, userID)
		if err != nil {
			log.ErrorContext(ctx, "list categories failed", slog.String("error", err.Error()))
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

// ExpenseCategoryHandler processes the category callback for the expense flow.
func ExpenseCategoryHandler(store *fsm.Store, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID

		categoryID, err := parseCategoryCallback(query.Data)
		if err != nil {
			log.WarnContext(ctx, "invalid category callback", slog.Any("error", err))
			sendError(ctx, b, query.Message.Message.Chat.ID)
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		if err := store.SetData(ctx, userID, "category", strconv.FormatInt(categoryID, 10)); err != nil {
			sendError(ctx, b, query.Message.Message.Chat.ID)
			return
		}
		if err := store.SetState(ctx, userID, fsm.StateExpenseWaitNote); err != nil {
			sendError(ctx, b, query.Message.Message.Chat.ID)
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

// ExpenseNoteHandler finalises the expense transaction.
func ExpenseNoteHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		note := update.Message.Text
		if note == "/skip" {
			note = ""
		}

		amountStr, err := store.GetData(ctx, userID, "amount")
		if err != nil {
			log.ErrorContext(ctx, "get fsm amount failed", slog.Int64("user_id", userID), slog.Any("error", err))
			_ = store.Clear(ctx, userID)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		categoryStr, err := store.GetData(ctx, userID, "category")
		if err != nil {
			log.ErrorContext(ctx, "get fsm category failed", slog.Int64("user_id", userID), slog.Any("error", err))
			_ = store.Clear(ctx, userID)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		cents, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil || cents <= 0 {
			_ = store.Clear(ctx, userID)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}
		catID, err := strconv.ParseInt(categoryStr, 10, 64)
		if err != nil || catID <= 0 {
			_ = store.Clear(ctx, userID)
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		tx, err := txSvc.AddExpense(ctx, userID, cents, catID, note)
		if err != nil {
			log.ErrorContext(ctx, "add expense failed", slog.String("error", err.Error()))
			sendError(ctx, b, update.Message.Chat.ID)
		} else {
			emoji := tx.CategoryEmoji
			if emoji == "" {
				emoji = "📦"
			}
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text: fmt.Sprintf("✅ Expense recorded: -%s %s %s",
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
