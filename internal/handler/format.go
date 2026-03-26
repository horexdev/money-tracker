package handler

import (
	"fmt"
	"strings"

	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/pkg/money"
)

// Menu button labels sent as plain text by the persistent reply keyboard.
const (
	btnExpense  = "💸 Expense"
	btnIncome   = "💰 Income"
	btnBalance  = "💳 Balance"
	btnHistory  = "📋 History"
	btnStats    = "📊 Stats"
	btnSettings = "⚙️ Settings"
	btnDevblog  = "📓 Devblog"
	btnCancel   = "❌ Cancel"
)

// mainMenuKeyboard returns the persistent reply keyboard shown in the bot's home state.
func mainMenuKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: btnExpense}, {Text: btnIncome}},
			{{Text: btnBalance}, {Text: btnHistory}},
			{{Text: btnStats}, {Text: btnSettings}},
			{{Text: btnDevblog}},
		},
		ResizeKeyboard: true,
	}
}

// buildDevblogListKeyboard creates pagination controls for the devblog list.
// Callback format: "dlog:list:{page}".
func buildDevblogListKeyboard(page, totalPages int) *models.InlineKeyboardMarkup {
	var row []models.InlineKeyboardButton

	if page > 1 {
		row = append(row, models.InlineKeyboardButton{
			Text: "‹", CallbackData: fmt.Sprintf("dlog:list:%d", page-1),
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
			Text: "›", CallbackData: fmt.Sprintf("dlog:list:%d", page+1),
		})
	} else {
		row = append(row, models.InlineKeyboardButton{
			Text: "›", CallbackData: "noop",
		})
	}

	return &models.InlineKeyboardMarkup{InlineKeyboard: [][]models.InlineKeyboardButton{row}}
}

// buildDevblogEntryKeyboard creates the back button for a single devblog entry view.
func buildDevblogEntryKeyboard(listPage int) *models.InlineKeyboardMarkup {
	return &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{
				{
					Text:         "◀ Back to list",
					CallbackData: fmt.Sprintf("dlog:list:%d", listPage),
				},
			},
		},
	}
}

// cancelKeyboard returns a single-row reply keyboard shown during active FSM flows.
func cancelKeyboard() *models.ReplyKeyboardMarkup {
	return &models.ReplyKeyboardMarkup{
		Keyboard: [][]models.KeyboardButton{
			{{Text: btnCancel}},
		},
		ResizeKeyboard: true,
	}
}

// escapeHTML escapes characters that have special meaning in Telegram HTML.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// formatSignedAmount returns a signed, HTML-wrapped amount string like "<code>-12.50</code>".
func formatSignedAmount(cents int64, isExpense bool) string {
	if isExpense {
		return fmt.Sprintf("<code>-%s</code>", money.FormatCents(cents))
	}
	return fmt.Sprintf("<code>+%s</code>", money.FormatCents(cents))
}

// formatNetAmount returns a signed net amount like "+12.50" or "-3.00".
func formatNetAmount(cents int64) string {
	if cents >= 0 {
		return "+" + money.FormatCents(cents)
	}
	return money.FormatCents(cents)
}

// categoryEmoji returns the emoji for a category, falling back to a default.
func categoryEmoji(emoji string) string {
	if emoji == "" {
		return "📦"
	}
	return emoji
}
