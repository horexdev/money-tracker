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
	exchangeSvc *service.ExchangeService,
	log *slog.Logger,
) {
	// Slash commands (backward compatible)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, StartHandler(log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, StartHandler(log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/cancel", bot.MatchTypeExact, CancelHandler(store, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/balance", bot.MatchTypeExact, BalanceHandler(txSvc, userSvc, exchangeSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/history", bot.MatchTypeExact, HistoryHandler(txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/addexpense", bot.MatchTypeExact, TransactionStartHandler(store, txSvc, userSvc, "expense", log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/addincome", bot.MatchTypeExact, TransactionStartHandler(store, txSvc, userSvc, "income", log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/stats", bot.MatchTypeExact, StatsStartHandler(log))
	b.RegisterHandler(bot.HandlerTypeMessageText, "/settings", bot.MatchTypeExact, SettingsHandler(userSvc, log))

	// Reply keyboard button text handlers
	b.RegisterHandler(bot.HandlerTypeMessageText, btnExpense, bot.MatchTypeExact, TransactionStartHandler(store, txSvc, userSvc, "expense", log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnIncome, bot.MatchTypeExact, TransactionStartHandler(store, txSvc, userSvc, "income", log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnBalance, bot.MatchTypeExact, BalanceHandler(txSvc, userSvc, exchangeSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnHistory, bot.MatchTypeExact, HistoryHandler(txSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnStats, bot.MatchTypeExact, StatsStartHandler(log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnSettings, bot.MatchTypeExact, SettingsHandler(userSvc, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnCancel, bot.MatchTypeExact, CancelHandler(store, log))
	b.RegisterHandler(bot.HandlerTypeMessageText, btnDevblog, bot.MatchTypeExact, DevblogHandler(log))

	// Callback queries
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "cat:", bot.MatchTypePrefix, TransactionCategoryHandler(store, txSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "period:", bot.MatchTypePrefix, StatsPeriodHandler(statsSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "flow:", bot.MatchTypePrefix, FlowCallbackHandler(store, txSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "hist:", bot.MatchTypePrefix, HistoryPageHandler(txSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "nav:", bot.MatchTypePrefix, NavHandler(store, txSvc, userSvc, statsSvc, exchangeSvc, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "settings:", bot.MatchTypePrefix, SettingsCallbackHandler(userSvc, store, log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "stats:", bot.MatchTypePrefix, StatsReselectHandler(log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "noop", bot.MatchTypeExact, NoopHandler(log))
	b.RegisterHandler(bot.HandlerTypeCallbackQueryData, "dlog:", bot.MatchTypePrefix, DevblogCallbackHandler(log))
}

// DefaultHandler is the fallback for non-command text messages.
// It dispatches based on the user's current FSM state.
func DefaultHandler(
	store *fsm.Store,
	txSvc *service.TransactionService,
	userSvc *service.UserService,
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
			sendErrorWithMenu(ctx, b, update.Message.Chat.ID, log)
			return
		}

		switch state {
		case fsm.StateExpenseWaitAmount, fsm.StateIncomeWaitAmount:
			TransactionAmountHandler(store, txSvc, log)(ctx, b, update)
		case fsm.StateExpenseWaitNote, fsm.StateIncomeWaitNote:
			TransactionNoteHandler(store, txSvc, log)(ctx, b, update)
		case fsm.StateCurrencySearch, fsm.StateDisplayCurrencySearch:
			CurrencySearchHandler(store, userSvc, log)(ctx, b, update)
		default:
			sendWithMainMenu(ctx, b, update.Message.Chat.ID,
				"I'm not sure what to do with that. Tap a button below or use /help.", log)
		}
	}
}

// buildCategoryKeyboard creates an inline keyboard from a list of categories (3 per row).
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
		// 3 buttons per row
		if (i+1)%3 == 0 || i == len(cats)-1 {
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
				{Text: "This Month", CallbackData: "period:month"},
			},
			{
				{Text: "Last Week", CallbackData: "period:lastweek"},
				{Text: "Last Month", CallbackData: "period:lastmonth"},
				{Text: "Last 3 Months", CallbackData: "period:3months"},
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
