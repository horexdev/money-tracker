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
	"github.com/horexdev/money-tracker/pkg/money"
)

// TransactionStartHandler initiates the add-expense or add-income flow.
func TransactionStartHandler(store *fsm.Store, txSvc *service.TransactionService, userSvc *service.UserService, txType string, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		startTransactionFlow(ctx, b, store, userSvc, update.Message.From.ID, update.Message.Chat.ID, txType, log)
	}
}

// startTransactionFlow is a shared entry point used by both text commands and nav callbacks.
func startTransactionFlow(ctx context.Context, b *bot.Bot, store *fsm.Store, userSvc *service.UserService, userID, chatID int64, txType string, log *slog.Logger) {
	var state fsm.State
	if txType == "expense" {
		state = fsm.StateExpenseWaitAmount
	} else {
		state = fsm.StateIncomeWaitAmount
	}

	if err := store.SetState(ctx, userID, state); err != nil {
		log.ErrorContext(ctx, "set state failed", slog.String("error", err.Error()))
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}
	if err := store.SetData(ctx, userID, "tx_type", txType); err != nil {
		log.ErrorContext(ctx, "set tx_type failed", slog.String("error", err.Error()))
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	// Store the user's current currency for this transaction.
	currencyCode := "USD"
	if userSvc != nil {
		if u, err := userSvc.GetByID(ctx, userID); err == nil {
			currencyCode = u.CurrencyCode
		}
	}
	if err := store.SetData(ctx, userID, "currency", currencyCode); err != nil {
		log.ErrorContext(ctx, "set currency failed", slog.String("error", err.Error()))
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	emoji := "💸"
	label := "Expense"
	if txType == "income" {
		emoji = "💰"
		label = "Income"
	}

	text := fmt.Sprintf("<b>%s New %s</b>\n\nEnter the amount:", emoji, label)
	sendWithCancel(ctx, b, chatID, text, log)
}

// TransactionAmountHandler processes the amount step.
func TransactionAmountHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := update.Message.Text

		cents, err := money.ParseCents(text)
		if err != nil {
			sendWithCancel(ctx, b, chatID, "That doesn't look like a valid amount. Try something like <b>12.50</b> or <b>300</b>", log)
			return
		}

		if err := store.SetData(ctx, userID, "amount", strconv.FormatInt(cents, 10)); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		txType, _ := store.GetData(ctx, userID, "tx_type")
		var state fsm.State
		if txType == "expense" {
			state = fsm.StateExpenseWaitCategory
		} else {
			state = fsm.StateIncomeWaitCategory
		}
		if err := store.SetState(ctx, userID, state); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		cats, err := txSvc.ListCategories(ctx, userID)
		if err != nil {
			log.ErrorContext(ctx, "list categories failed", slog.String("error", err.Error()))
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		sendWithInline(ctx, b, chatID, "<b>📂 Choose a category:</b>", buildCategoryKeyboard(cats), log)
	}
}

// TransactionCategoryHandler processes the category callback.
func TransactionCategoryHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID
		chatID := query.Message.Message.Chat.ID

		categoryID, err := parseCategoryCallback(query.Data)
		if err != nil {
			log.WarnContext(ctx, "invalid category callback", slog.Any("error", err))
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		// Fetch category to cache name/emoji for confirmation card.
		cats, err := txSvc.ListCategories(ctx, userID)
		if err != nil {
			log.ErrorContext(ctx, "list categories failed", slog.String("error", err.Error()))
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}
		var catName, catEmoji string
		for _, c := range cats {
			if c.ID == categoryID {
				catName = c.Name
				catEmoji = c.Emoji
				break
			}
		}

		if err := store.SetData(ctx, userID, "category", strconv.FormatInt(categoryID, 10)); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}
		if err := store.SetData(ctx, userID, "category_name", catName); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}
		if err := store.SetData(ctx, userID, "category_emoji", catEmoji); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		txType, _ := store.GetData(ctx, userID, "tx_type")
		var state fsm.State
		if txType == "expense" {
			state = fsm.StateExpenseWaitNote
		} else {
			state = fsm.StateIncomeWaitNote
		}
		if err := store.SetState(ctx, userID, state); err != nil {
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		skipKb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "⏭ Skip", CallbackData: "flow:skip_note"}},
			},
		}
		sendWithInline(ctx, b, chatID, "<b>📝 Add a note:</b>", skipKb, log)
	}
}

// TransactionNoteHandler processes the note text or /skip.
func TransactionNoteHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		note := update.Message.Text
		if note == "/skip" {
			note = ""
		}
		showConfirmation(ctx, b, store, userID, chatID, note, log)
	}
}

// FlowCallbackHandler handles all "flow:*" callbacks.
func FlowCallbackHandler(store *fsm.Store, txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}
		userID := query.From.ID
		chatID := query.Message.Message.Chat.ID

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		action := strings.TrimPrefix(query.Data, "flow:")

		switch action {
		case "skip_note":
			showConfirmation(ctx, b, store, userID, chatID, "", log)

		case "confirm":
			confirmTransaction(ctx, b, store, txSvc, userID, chatID, log)

		case "cancel":
			if err := store.Clear(ctx, userID); err != nil {
				log.ErrorContext(ctx, "failed to clear FSM", slog.String("error", err.Error()))
			}
			sendWithMainMenu(ctx, b, chatID, "Cancelled. What would you like to do?", log)
		}
	}
}

// showConfirmation displays the confirmation card before saving the transaction.
func showConfirmation(ctx context.Context, b *bot.Bot, store *fsm.Store, userID, chatID int64, note string, log *slog.Logger) {
	if err := store.SetData(ctx, userID, "note", note); err != nil {
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	txType, _ := store.GetData(ctx, userID, "tx_type")
	amountStr, _ := store.GetData(ctx, userID, "amount")
	catName, _ := store.GetData(ctx, userID, "category_name")
	catEmoji, _ := store.GetData(ctx, userID, "category_emoji")
	currencyCode, _ := store.GetData(ctx, userID, "currency")

	cents, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		_ = store.Clear(ctx, userID)
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	var state fsm.State
	label := "expense"
	if txType == "income" {
		state = fsm.StateIncomeWaitConfirm
		label = "income"
	} else {
		state = fsm.StateExpenseWaitConfirm
	}
	if err := store.SetState(ctx, userID, state); err != nil {
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	emoji := categoryEmoji(catEmoji)
	noteDisplay := "<i>none</i>"
	if note != "" {
		noteDisplay = escapeHTML(note)
	}

	sym := money.CurrencySymbol(currencyCode)
	text := fmt.Sprintf(
		"<b>Confirm %s?</b>\n\n"+
			"💰 Amount:    <code>%s %s</code>\n"+
			"📂 Category:  %s %s\n"+
			"📝 Note:      %s",
		label,
		sym, money.FormatCents(cents),
		emoji, escapeHTML(catName),
		noteDisplay,
	)

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "✅ Confirm", CallbackData: "flow:confirm"},
				{Text: "❌ Cancel", CallbackData: "flow:cancel"},
			},
		},
	}
	sendWithInline(ctx, b, chatID, text, kb, log)
}

// confirmTransaction saves the transaction and shows a success card.
func confirmTransaction(ctx context.Context, b *bot.Bot, store *fsm.Store, txSvc *service.TransactionService, userID, chatID int64, log *slog.Logger) {
	txType, _ := store.GetData(ctx, userID, "tx_type")
	amountStr, _ := store.GetData(ctx, userID, "amount")
	categoryStr, _ := store.GetData(ctx, userID, "category")
	note, _ := store.GetData(ctx, userID, "note")
	currencyCode, _ := store.GetData(ctx, userID, "currency")
	if currencyCode == "" {
		currencyCode = "USD"
	}

	cents, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil || cents <= 0 {
		_ = store.Clear(ctx, userID)
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}
	catID, err := strconv.ParseInt(categoryStr, 10, 64)
	if err != nil || catID <= 0 {
		_ = store.Clear(ctx, userID)
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	var tx *domain.Transaction
	if txType == "income" {
		tx, err = txSvc.AddIncome(ctx, userID, cents, catID, note, currencyCode)
	} else {
		tx, err = txSvc.AddExpense(ctx, userID, cents, catID, note, currencyCode)
	}
	if err != nil {
		log.ErrorContext(ctx, "add transaction failed", slog.String("error", err.Error()))
		_ = store.Clear(ctx, userID)
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	if err := store.Clear(ctx, userID); err != nil {
		log.ErrorContext(ctx, "failed to clear FSM", slog.String("error", err.Error()))
	}

	emoji := categoryEmoji(tx.CategoryEmoji)
	isExpense := tx.Type == domain.TransactionTypeExpense
	sign := formatSignedAmount(tx.AmountCents, isExpense)

	label := "Income"
	if isExpense {
		label = "Expense"
	}

	var noteBlock string
	if tx.Note != "" {
		noteBlock = "\n" + escapeHTML(tx.Note)
	}

	text := fmt.Sprintf("<b>✅ %s recorded!</b>\n\n%s  %s %s%s",
		label, sign, emoji, escapeHTML(tx.CategoryName), noteBlock)

	// Quick follow-up buttons
	var anotherLabel, anotherData string
	if isExpense {
		anotherLabel = "💸 Another Expense"
		anotherData = "nav:expense"
	} else {
		anotherLabel = "💰 Another Income"
		anotherData = "nav:income"
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: anotherLabel, CallbackData: anotherData},
				{Text: "💳 Balance", CallbackData: "nav:balance"},
			},
		},
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}); err != nil {
		log.ErrorContext(ctx, "failed to send success message", slog.String("error", err.Error()))
	}

	// Restore main menu keyboard with a silent message isn't ideal.
	// Instead, send the success message with inline buttons and the main menu keyboard
	// is restored on next interaction. We ensure it's there by always including it
	// in response to any menu button or command.
}

// sendErrorWithMenu sends a generic error message with the main menu keyboard.
func sendErrorWithMenu(ctx context.Context, b *bot.Bot, chatID int64, log *slog.Logger) {
	sendWithMainMenu(ctx, b, chatID, "Something went wrong. Please try again.", log)
}
