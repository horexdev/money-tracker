package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/fsm"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

// SettingsHandler shows the settings screen.
func SettingsHandler(userSvc *service.UserService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil || update.Message.From == nil {
			return
		}
		sendSettings(ctx, b, userSvc, update.Message.From.ID, update.Message.Chat.ID, log)
	}
}

// SettingsCallbackHandler handles all "settings:*" callbacks.
func SettingsCallbackHandler(userSvc *service.UserService, store *fsm.Store, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		userID := query.From.ID
		chatID := query.Message.Message.Chat.ID
		messageID := query.Message.Message.ID
		action := strings.TrimPrefix(query.Data, "settings:")

		switch {
		case action == "base_currency":
			// Enter FSM state for currency search.
			if err := store.SetState(ctx, userID, fsm.StateCurrencySearch); err != nil {
				log.ErrorContext(ctx, "failed to set FSM state", slog.String("error", err.Error()))
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			sendWithCancel(ctx, b, chatID,
				"<b>🔄 Change Base Currency</b>\n\nType a currency code or name (e.g. <b>EUR</b>, <b>Ruble</b>, <b>Yen</b>):", log)

		case strings.HasPrefix(action, "setbase:"):
			code := strings.TrimPrefix(action, "setbase:")
			if err := store.Clear(ctx, userID); err != nil {
				log.ErrorContext(ctx, "failed to clear FSM", slog.String("error", err.Error()))
			}
			_, err := userSvc.UpdateCurrency(ctx, userID, code)
			if err != nil {
				log.ErrorContext(ctx, "failed to update currency", slog.String("error", err.Error()))
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			sym := money.CurrencySymbol(code)
			text := fmt.Sprintf("✅ Base currency set to <b>%s %s</b>", sym, code)
			sendWithMainMenu(ctx, b, chatID, text, log)

		case action == "display":
			sendDisplayCurrenciesCard(ctx, b, userSvc, userID, chatID, messageID, log)

		case action == "add_display":
			if err := store.SetState(ctx, userID, fsm.StateDisplayCurrencySearch); err != nil {
				log.ErrorContext(ctx, "failed to set FSM state", slog.String("error", err.Error()))
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			sendWithCancel(ctx, b, chatID,
				"<b>➕ Add Display Currency</b>\n\nType a currency code or name (e.g. <b>EUR</b>, <b>Pound</b>):", log)

		case strings.HasPrefix(action, "adddisplay:"):
			code := strings.TrimPrefix(action, "adddisplay:")
			if err := store.Clear(ctx, userID); err != nil {
				log.ErrorContext(ctx, "failed to clear FSM", slog.String("error", err.Error()))
			}
			u, err := userSvc.GetByID(ctx, userID)
			if err != nil {
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			// Check if already present or at limit.
			for _, c := range u.DisplayCurrencies {
				if c == code {
					sendWithMainMenu(ctx, b, chatID, fmt.Sprintf("%s is already in your display currencies.", code), log)
					return
				}
			}
			if len(u.DisplayCurrencies) >= 3 {
				sendWithMainMenu(ctx, b, chatID, "You can have at most 3 display currencies. Remove one first.", log)
				return
			}
			newCodes := append(u.DisplayCurrencies, code)
			if _, err := userSvc.UpdateDisplayCurrencies(ctx, userID, newCodes); err != nil {
				log.ErrorContext(ctx, "failed to add display currency", slog.String("error", err.Error()))
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			sym := money.CurrencySymbol(code)
			sendWithMainMenu(ctx, b, chatID, fmt.Sprintf("✅ Added <b>%s %s</b> to display currencies.", sym, code), log)

		case strings.HasPrefix(action, "remdisplay:"):
			code := strings.TrimPrefix(action, "remdisplay:")
			u, err := userSvc.GetByID(ctx, userID)
			if err != nil {
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			newCodes := make([]string, 0, len(u.DisplayCurrencies))
			for _, c := range u.DisplayCurrencies {
				if c != code {
					newCodes = append(newCodes, c)
				}
			}
			if _, err := userSvc.UpdateDisplayCurrencies(ctx, userID, newCodes); err != nil {
				log.ErrorContext(ctx, "failed to remove display currency", slog.String("error", err.Error()))
				sendErrorWithMenu(ctx, b, chatID, log)
				return
			}
			sendDisplayCurrenciesCard(ctx, b, userSvc, userID, chatID, messageID, log)

		case action == "back":
			sendSettingsCard(ctx, b, userSvc, userID, chatID, messageID, log)
		}
	}
}

// CurrencySearchHandler handles text input during StateCurrencySearch.
func CurrencySearchHandler(store *fsm.Store, userSvc *service.UserService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID

		state, _ := store.GetState(ctx, userID)
		callbackPrefix := "settings:setbase:"
		if state == fsm.StateDisplayCurrencySearch {
			callbackPrefix = "settings:adddisplay:"
		}

		query := update.Message.Text
		results := domain.SearchCurrencies(query)
		if len(results) == 0 {
			sendWithCancel(ctx, b, chatID,
				fmt.Sprintf("No currencies found for <b>%s</b>. Try again:", escapeHTML(query)), log)
			return
		}

		var rows [][]models.InlineKeyboardButton
		for _, c := range results {
			rows = append(rows, []models.InlineKeyboardButton{
				{
					Text:         fmt.Sprintf("%s %s — %s", c.Symbol, c.Code, c.Name),
					CallbackData: callbackPrefix + c.Code,
				},
			})
		}
		kb := &models.InlineKeyboardMarkup{InlineKeyboard: rows}
		sendWithInline(ctx, b, chatID,
			fmt.Sprintf("Results for <b>%s</b>:", escapeHTML(query)), kb, log)
	}
}

// sendSettings sends the settings card as a new message.
func sendSettings(ctx context.Context, b *bot.Bot, userSvc *service.UserService, userID, chatID int64, log *slog.Logger) {
	u, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user", slog.String("error", err.Error()))
		sendErrorWithMenu(ctx, b, chatID, log)
		return
	}

	text, kb := buildSettingsTextAndKeyboard(u)
	if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}); err != nil {
		log.ErrorContext(ctx, "failed to send settings", slog.String("error", err.Error()))
	}
}

// sendSettingsCard edits an existing message to show the settings card.
func sendSettingsCard(ctx context.Context, b *bot.Bot, userSvc *service.UserService, userID, chatID int64, messageID int, log *slog.Logger) {
	u, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user", slog.String("error", err.Error()))
		return
	}

	text, kb := buildSettingsTextAndKeyboard(u)
	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}); err != nil {
		log.ErrorContext(ctx, "failed to edit settings message", slog.String("error", err.Error()))
	}
}

// buildSettingsTextAndKeyboard builds the settings card text and keyboard.
func buildSettingsTextAndKeyboard(u *domain.User) (string, *models.InlineKeyboardMarkup) {
	sym := money.CurrencySymbol(u.CurrencyCode)

	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>⚙️ Settings</b>\n\n")
	fmt.Fprintf(&sb, "💱 Base currency: <b>%s %s</b>\n", sym, u.CurrencyCode)

	if len(u.DisplayCurrencies) > 0 {
		sb.WriteString("📊 Display currencies: ")
		for i, c := range u.DisplayCurrencies {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(&sb, "<b>%s %s</b>", money.CurrencySymbol(c), c)
		}
		sb.WriteString("\n")
	} else {
		sb.WriteString("📊 Display currencies: <i>none</i>\n")
	}

	kb := &models.InlineKeyboardMarkup{
		InlineKeyboard: [][]models.InlineKeyboardButton{
			{{Text: "🔄 Change Base Currency", CallbackData: "settings:base_currency"}},
			{{Text: "📊 Display Currencies", CallbackData: "settings:display"}},
		},
	}

	return sb.String(), kb
}

// sendDisplayCurrenciesCard shows current display currencies with remove/add buttons.
func sendDisplayCurrenciesCard(ctx context.Context, b *bot.Bot, userSvc *service.UserService, userID, chatID int64, messageID int, log *slog.Logger) {
	u, err := userSvc.GetByID(ctx, userID)
	if err != nil {
		log.ErrorContext(ctx, "failed to get user", slog.String("error", err.Error()))
		return
	}

	var sb strings.Builder
	sb.WriteString("<b>📊 Display Currencies</b>\n\n")
	sb.WriteString("Converted amounts will be shown in these currencies alongside your base currency.\n\n")

	var rows [][]models.InlineKeyboardButton

	if len(u.DisplayCurrencies) > 0 {
		for _, c := range u.DisplayCurrencies {
			sym := money.CurrencySymbol(c)
			rows = append(rows, []models.InlineKeyboardButton{
				{Text: fmt.Sprintf("❌ %s %s", sym, c), CallbackData: "settings:remdisplay:" + c},
			})
		}
	} else {
		sb.WriteString("<i>No display currencies set.</i>\n")
	}

	if len(u.DisplayCurrencies) < 3 {
		rows = append(rows, []models.InlineKeyboardButton{
			{Text: "➕ Add Currency", CallbackData: "settings:add_display"},
		})
	}

	rows = append(rows, []models.InlineKeyboardButton{
		{Text: "◀ Back", CallbackData: "settings:back"},
	})

	kb := &models.InlineKeyboardMarkup{InlineKeyboard: rows}

	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        sb.String(),
		ParseMode:   models.ParseModeHTML,
		ReplyMarkup: kb,
	}); err != nil {
		log.ErrorContext(ctx, "failed to edit display currencies", slog.String("error", err.Error()))
	}
}
