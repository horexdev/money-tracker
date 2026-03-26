package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/service"
)

// RegisterAll attaches all command and callback handlers to the bot instance.
func RegisterAll(
	b *bot.Bot,
	store *fsm.Store,
	userSvc *service.UserService,
	txSvc *service.TransactionService,
	statsSvc *service.StatsService,
	log *slog.Logger,
) {
	// Commands
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, StartHandler())
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, StartHandler())
	b.RegisterHandler(bot.HandlerTypeMessageText, "/cancel", bot.MatchTypeExact, CancelHandler(store, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, BalanceHandler(txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/history", bot.MatchTypeExact, HistoryHandler(txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/addexpense", bot.MatchTypeExact, ExpenseStartHandler(store, txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/addincome", bot.MatchTypeExact, IncomeStartHandler(store, txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypeExact, StatsStartHandler(store, log))

	// Callback queries
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "cat:", bot.MatchTypePrefix, dispatchCategoryCallback(store, txSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "period:", bot.MatchTypePrefix, StatsPeriodHandler(store, statsSvc, log))
}

// DefaultHandler is the fallback for non-command text messages.
// It dispatches based on the user's current FSM state.
func DefaultHandler(
	store *fsm.Store,
	txSvc *service.TransactionService,
	log *slog.Logger,
) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID

		state, err := store.GetState(ctx, userID)
		if err != nil {
			log.ErrorContext(ctx, "get FSM state failed", slog.String("error", err.Error()))
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		switch state {
		case fsm.StateExpenseWaitAmount:
			ExpenseAmountHandler(store, txSvc, log)(ctx, b, update)
		case fsm.StateExpenseWaitNote:
			ExpenseNoteHandler(store, txSvc, log)(ctx, b, update)
		case fsm.StateIncomeWaitAmount:
			IncomeAmountHandler(store, txSvc, log)(ctx, b, update)
		case fsm.StateIncomeWaitNote:
			IncomeNoteHandler(store, txSvc, log)(ctx, b, update)
		default:
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Use /help to see available commands.",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
		}
	}
}

// dispatchCategoryCallback routes category callbacks to the correct flow handler.
func dispatchCategoryCallback(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.CallbackQuery == nil {
			return
		}
		userID := update.CallbackQuery.From.ID

		state, err := store.GetState(ctx, userID)
		if err != nil {
			return
		}

		switch state {
		case fsm.StateExpenseWaitCategory:
			ExpenseCategoryHandler(store, log)(ctx, b, update)
		case fsm.StateIncomeWaitCategory:
			IncomeCategoryHandler(store, log)(ctx, b, update)
		default:
			if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "No active flow. Use /addexpense or /addincome.",
			}); err != nil {
				log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
			}
		}
	}
}

// buildCategoryKeyboard creates an inline keyboard from a list of categories.
func buildCategoryKeyboard(cats []*domain.Category) *models.InlineKeyboardMarkup {
	var rows [][]models.InlineKeyboardButton
	var row []models.InlineKeyboardButton

	for i, cat := range cats {
		label := cat.Name
		if cat.Emoji != "" {
			label = cat.Emoji + " " + cat.Name
		}
		row = append(row, models.InlineKeyboardButton{
			Text:         label,
			CallbackData: fmt.Sprintf("cat:%d", cat.ID),
		})
		// 2 buttons per row
		if (i+1)%2 == 0 || i == len(cats)-1 {
			rows = append(rows, row)
			row = nil
		}
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// buildPeriodKeyboard creates an inline keyboard for period selection.
func buildPeriodKeyboard() *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "Today", CallbackData: "period:today"},
				{Text: "This Week", CallbackData: "period:week"},
			},
			{
				{Text: "This Month", CallbackData: "period:month"},
				{Text: "Last Month", CallbackData: "period:lastmonth"},
			},
		},
	}
}

// parseCategoryCallback extracts the category ID from a "cat:{id}" callback.
// Returns an error if the data is malformed or the ID is not positive.
func parseCategoryCallback(data string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimPrefix(data, "cat:"), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid category callback data %q", data)
	}
	return id, nil
}

// parsePeriodCallback extracts the period name from a "period:{name}" callback.
func parsePeriodCallback(data string) string {
	return strings.TrimPrefix(data, "period:")
}

// sendError sends a generic error message to the chat.
func sendError(ctx context.Context, b *bot.Bot, chatID int64) {
	_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   "❌ Something went wrong. Please try again or use /cancel.",
	})
}
