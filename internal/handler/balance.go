package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

// BalanceHandler shows the user's current net balance.
func BalanceHandler(txSvc *service.TransactionService, userSvc *service.UserService, exchangeSvc *service.ExchangeService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		sendBalance(ctx, b, txSvc, userSvc, exchangeSvc, update.Message.From.ID, update.Message.Chat.ID, log)
	}
}

// sendBalance renders the balance card. Used by both the text command and nav callbacks.
func sendBalance(ctx context.Context, b *bot.Bot, txSvc *service.TransactionService, userSvc *service.UserService, exchangeSvc *service.ExchangeService, userID, chatID int64, log *slog.Logger) {
	// Fetch per-currency balances.
	balances, err := txSvc.GetBalanceByCurrency(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get balance",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	if len(balances) == 0 {
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "💸 Add Expense", CallbackData: "nav:expense"},
					{Text: "💰 Add Income", CallbackData: "nav:income"},
				},
			},
		}
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    chatID,
			Text:      "<b>💳 Balance</b>\n\nYour balance is empty. Start tracking by\nrecording your first transaction!",
			ParseMode: models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send balance", slog.String("error", err.Error()))
		}
		return
	}

	// Get user for base currency and display currencies.
	u, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user", slog.String("error", err.Error()))
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	baseCurrency := u.CurrencyCode
	baseSym := money.CurrencySymbol(baseCurrency)

	// Convert all per-currency balances to base currency.
	var totalIncome, totalExpense int64
	for _, bal := range balances {
		if bal.CurrencyCode == baseCurrency {
			totalIncome += bal.IncomeCents
			totalExpense += bal.ExpenseCents
		} else {
			inc, err := exchangeSvc.Convert(ctx, bal.IncomeCents, strings.TrimSpace(bal.CurrencyCode), baseCurrency)
			if err != nil {
				// If conversion fails, use raw amounts as fallback.
				totalIncome += bal.IncomeCents
				totalExpense += bal.ExpenseCents
				continue
			}
			exp, err := exchangeSvc.Convert(ctx, bal.ExpenseCents, strings.TrimSpace(bal.CurrencyCode), baseCurrency)
			if err != nil {
				totalIncome += bal.IncomeCents
				totalExpense += bal.ExpenseCents
				continue
			}
			totalIncome += inc
			totalExpense += exp
		}
	}

	net := totalIncome - totalExpense

	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>💳 Balance</b>  <i>(%s)</i>\n\n", baseCurrency)
	fmt.Fprintf(&sb, "Income:    <code>+%s %s</code>\n", baseSym, money.FormatCents(totalIncome))
	fmt.Fprintf(&sb, "Expense:   <code>-%s %s</code>\n", baseSym, money.FormatCents(totalExpense))
	sb.WriteString("─────────────────\n")
	fmt.Fprintf(&sb, "Net:       <code>%s %s</code>", baseSym, formatNetAmount(net))

	// Show display currency conversions if configured.
	if len(u.DisplayCurrencies) > 0 {
		converted, err := exchangeSvc.ConvertMulti(ctx, net, baseCurrency, u.DisplayCurrencies)
		if err == nil && len(converted) > 0 {
			sb.WriteString("\n\n")
			for _, dc := range u.DisplayCurrencies {
				if amt, ok := converted[dc]; ok {
					dcSym := money.CurrencySymbol(dc)
					fmt.Fprintf(&sb, "  ≈ <code>%s %s</code> %s\n", dcSym, formatNetAmount(amt), dc)
				}
			}
		}
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{Text: "💸 Expense", CallbackData: "nav:expense"},
				{Text: "💰 Income", CallbackData: "nav:income"},
			},
			{
				{Text: "📊 View Stats", CallbackData: "nav:stats"},
			},
		},
	}

	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        sb.String(),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}); err != nil {
		log.ErrorContext(ctx, "failed to send balance", slog.String("error", err.Error()))
	}
}
