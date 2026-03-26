package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/internal/domain"
	"github.com/horexdev/money-tracker/internal/service"
	"github.com/horexdev/money-tracker/pkg/money"
)

const barWidth = 10

// StatsStartHandler begins the stats flow by showing the period picker.
func StatsStartHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		sendStatsMenu(ctx, b, update.Message.Chat.ID, log)
	}
}

// sendStatsMenu sends the period picker inline keyboard.
func sendStatsMenu(ctx context.Context, b *bot.Bot, chatID int64, log *slog.Logger) {
	sendWithInline(ctx, b, chatID, "<b>📊 Choose a period:</b>", buildPeriodKeyboard(), log)
}

// StatsPeriodHandler handles the period callback and renders the stats report.
func StatsPeriodHandler(statsSvc *service.StatsService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID
		chatID := query.Message.Message.Chat.ID
		messageID := query.Message.Message.ID

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		period := parsePeriodCallback(query.Data)
		from, to, err := service.PeriodRange(period)
		if err != nil {
			if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:    chatID,
				MessageID: messageID,
				Text:      "Unknown period. Please try again.",
				ParseMode: models.ParseModeHTML,
			}); err != nil {
				log.ErrorContext(ctx, "failed to edit message", slog.String("error", err.Error()))
			}
			return
		}

		stats, err := statsSvc.ByCategory(ctx, userID, from, to)
		if err != nil {
			log.ErrorContext(ctx, "stats query failed", slog.String("error", err.Error()))
			sendErrorWithMenu(ctx, b, chatID, log)
			return
		}

		if len(stats) == 0 {
			kb := &models.InlineKeyboardMarkup{
				InlineKeyboard: [][]models.InlineKeyboardButton{
					{{Text: "◀ Change Period", CallbackData: "stats:reselect"}},
				},
			}
			if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
				ChatID:      chatID,
				MessageID:   messageID,
				Text:        fmt.Sprintf("<b>📊 Stats — %s</b>\n\nNo transactions found for this period.", periodLabel(period)),
				ParseMode:   models.ParseModeHTML,
				ReplyMarkup: kb,
			}); err != nil {
				log.ErrorContext(ctx, "failed to edit message", slog.String("error", err.Error()))
			}
			return
		}

		text := formatStatsHTML(stats, period)
		kb := &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{
				{{Text: "◀ Change Period", CallbackData: "stats:reselect"}},
			},
		}

		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   messageID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send stats", slog.String("error", err.Error()))
		}
	}
}

// StatsReselectHandler handles "stats:reselect" callback to re-show the period picker.
func StatsReselectHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}
		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      query.Message.Message.Chat.ID,
			MessageID:   query.Message.Message.ID,
			Text:        "<b>📊 Choose a period:</b>",
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: buildPeriodKeyboard(),
		}); err != nil {
			log.ErrorContext(ctx, "failed to edit message", slog.String("error", err.Error()))
		}
	}
}

// formatStatsHTML renders the stats report with Unicode bar charts.
func formatStatsHTML(stats []domain.CategoryStat, period string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>📊 Stats — %s</b>\n", periodLabel(period))

	// Separate expense and income stats.
	var expenses, incomes []domain.CategoryStat
	var totalExpense, totalIncome int64
	for _, s := range stats {
		if s.Type == domain.TransactionTypeIncome {
			incomes = append(incomes, s)
			totalIncome += s.TotalCents
		} else {
			expenses = append(expenses, s)
			totalExpense += s.TotalCents
		}
	}

	if len(expenses) > 0 {
		fmt.Fprintf(&sb, "\n<b>EXPENSES</b>  <code>-%s</code>\n\n", money.FormatCents(totalExpense))
		for _, s := range expenses {
			emoji := categoryEmoji(s.CategoryEmoji)
			pct := 0
			if totalExpense > 0 {
				pct = int(s.TotalCents * 100 / totalExpense)
			}
			fmt.Fprintf(&sb, "%s %s  <code>-%s</code>  %d%%\n",
				emoji, escapeHTML(s.CategoryName), money.FormatCents(s.TotalCents), pct)
			sb.WriteString(renderBar(pct))
			sb.WriteString("\n\n")
		}
	}

	if len(incomes) > 0 {
		fmt.Fprintf(&sb, "<b>INCOME</b>  <code>+%s</code>\n\n", money.FormatCents(totalIncome))
		for _, s := range incomes {
			emoji := categoryEmoji(s.CategoryEmoji)
			pct := 0
			if totalIncome > 0 {
				pct = int(s.TotalCents * 100 / totalIncome)
			}
			fmt.Fprintf(&sb, "%s %s  <code>+%s</code>  %d%%\n",
				emoji, escapeHTML(s.CategoryName), money.FormatCents(s.TotalCents), pct)
			sb.WriteString(renderBar(pct))
			sb.WriteString("\n\n")
		}
	}

	sb.WriteString("─────────────────\n")
	net := totalIncome - totalExpense
	fmt.Fprintf(&sb, "Net: <code>%s</code>", formatNetAmount(net))

	return sb.String()
}

// renderBar generates an emoji bar chart line visible on both light and dark themes.
func renderBar(pct int) string {
	filled := pct * barWidth / 100
	if filled < 0 {
		filled = 0
	}
	if filled > barWidth {
		filled = barWidth
	}
	empty := barWidth - filled
	return strings.Repeat("🟩", filled) + strings.Repeat("⬜", empty)
}

// periodLabel returns a human-readable label for a period key.
func periodLabel(period string) string {
	labels := map[string]string{
		"today":     "Today",
		"week":      "This Week",
		"month":     "This Month",
		"lastweek":  "Last Week",
		"lastmonth": "Last Month",
		"3months":   "Last 3 Months",
	}
	if l, ok := labels[period]; ok {
		return l
	}
	return period
}
