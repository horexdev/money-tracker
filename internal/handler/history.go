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
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

const historyPageSize = 5

// HistoryHandler shows the first page of the user's transactions.
func HistoryHandler(txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		text, kb := buildHistoryPage(ctx, txSvc, userID, 1, log)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send history", slog.String("error", err.Error()))
		}
	}
}

// HistoryPageHandler handles pagination callbacks for history ("hist:{page}").
func HistoryPageHandler(txSvc *service.TransactionService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		pageStr := strings.TrimPrefix(query.Data, "hist:")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		userID := query.From.ID
		text, kb := buildHistoryPage(ctx, txSvc, userID, page, log)

		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      query.Message.Message.Chat.ID,
			MessageID:   query.Message.Message.ID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "failed to edit history page", slog.String("error", err.Error()))
		}
	}
}

// buildHistoryPage renders the HTML text and keyboard for a given page.
func buildHistoryPage(ctx context.Context, txSvc *service.TransactionService, userID int64, page int, log *slog.Logger) (string, *models.InlineKeyboardMarkup) {
	txs, totalPages, err := txSvc.ListPaged(ctx, userID, page, historyPageSize)
	if err != nil {
		log.ErrorContext(ctx, "failed to list transactions",
			slog.Int64("user_id", userID),
			slog.String("error", err.Error()),
		)
		return "Something went wrong. Please try again.", nil
	}

	if len(txs) == 0 {
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{
					{Text: "💸 Add Expense", CallbackData: "nav:expense"},
					{Text: "💰 Add Income", CallbackData: "nav:income"},
				},
			},
		}
		return "<b>📋 Transactions</b>\n\nNo transactions yet. Your history will\nappear here once you start tracking!", kb
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>📋 Transactions</b>  <i>Page %d/%d</i>\n\n", page, totalPages)

	for _, tx := range txs {
		emoji := categoryEmoji(tx.CategoryEmoji)
		isExpense := tx.Type == domain.TransactionTypeExpense
		sign := formatSignedAmount(tx.AmountCents, isExpense)
		date := tx.CreatedAt.Format("02 Jan")
		cur := strings.TrimSpace(tx.CurrencyCode)
		sym := money.CurrencySymbol(cur)

		fmt.Fprintf(&sb, "%s  %s %s  %s%s\n", date, emoji, escapeHTML(tx.CategoryName), sym, sign)
		if tx.Note != "" {
			fmt.Fprintf(&sb, "         <i>%s</i>\n", escapeHTML(tx.Note))
		}
	}

	kb := buildPaginationKeyboard(page, totalPages)
	return sb.String(), kb
}

// buildPaginationKeyboard creates the navigation row for history.
func buildPaginationKeyboard(page, totalPages int) *models.InlineKeyboardMarkup {
	var row []models.InlineKeyboardButton

	if page > 1 {
		row = append(row, models.InlineKeyboardButton{
			Text: "‹", CallbackData: fmt.Sprintf("hist:%d", page-1),
		})
	} else {
		row = append(row, models.InlineKeyboardButton{
			Text: "‹", CallbackData: "noop",
		})
	}

	row = append(row, models.InlineKeyboardButton{
		Text:         fmt.Sprintf("%d / %d", page, totalPages),
		CallbackData: "noop",
	})

	if page < totalPages {
		row = append(row, models.InlineKeyboardButton{
			Text: "›", CallbackData: fmt.Sprintf("hist:%d", page+1),
		})
	} else {
		row = append(row, models.InlineKeyboardButton{
			Text: "›", CallbackData: "noop",
		})
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{row}}
}

