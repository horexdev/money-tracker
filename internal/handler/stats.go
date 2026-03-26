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

// StatsStartHandler begins the stats flow by asking the user to pick a period.
func StatsStartHandler(store *fsm.Store, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		if err := store.SetState(ctx, userID, fsm.StateStatsWaitPeriod); err != nil {
			sendError(ctx, b, update.Message.Chat.ID)
			return
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "📊 Choose a period:",
			ReplyMarkup: buildPeriodKeyboard(),
		}); err != nil {
			log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
		}
	}
}

// StatsPeriodHandler handles the period callback and renders the stats report.
func StatsPeriodHandler(store *fsm.Store, statsSvc *service.StatsService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID}); err != nil {
			log.ErrorContext(ctx, "failed to answer callback", slog.String("error", err.Error()))
		}

		period := parsePeriodCallback(query.Data)
		from, to, err := service.PeriodRange(period)
		if err != nil {
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: query.Message.Message.Chat.ID,
				Text:   "❌ Unknown period.",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
			if err := store.Clear(ctx, userID); err != nil {
				log.ErrorContext(ctx, "failed to clear FSM state", slog.String("error", err.Error()))
			}
			return
		}

		stats, err := statsSvc.ByCategory(ctx, userID, from, to)
		if err != nil {
			log.ErrorContext(ctx, "stats query failed", slog.String("error", err.Error()))
			sendErrorCallback(ctx, b, query.Message.Message.Chat.ID)
			if err := store.Clear(ctx, userID); err != nil {
				log.ErrorContext(ctx, "failed to clear FSM state", slog.String("error", err.Error()))
			}
			return
		}

		if err := store.Clear(ctx, userID); err != nil {
			log.ErrorContext(ctx, "failed to clear FSM state", slog.String("error", err.Error()))
		}

		if len(stats) == 0 {
			if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: query.Message.Message.Chat.ID,
				Text:   "No transactions found for this period.",
			}); err != nil {
				log.ErrorContext(ctx, "failed to send message", slog.String("error", err.Error()))
			}
			return
		}

		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    query.Message.Message.Chat.ID,
			Text:      formatStats(stats, period),
			ParseMode: models.ParseModeMarkdown,
		}); err != nil {
			log.ErrorContext(ctx, "failed to send stats", slog.String("error", err.Error()))
		}
	}
}

func formatStats(stats []domain.CategoryStat, period string) string {
	var sb strings.Builder
	periodLabels := map[string]string{
		"today":     "Today",
		"week":      "This Week",
		"month":     "This Month",
		"lastmonth": "Last Month",
	}
	label := periodLabels[period]
	if label == "" {
		label = period
	}

	fmt.Fprintf(&sb, "📊 *Stats — %s*\n\n", label)

	var totalExpense, totalIncome int64
	for _, s := range stats {
		emoji := s.CategoryEmoji
		if emoji == "" {
			emoji = "📦"
		}
		sign := "-"
		if s.Type == domain.TransactionTypeIncome {
			sign = "+"
			totalIncome += s.TotalCents
		} else {
			totalExpense += s.TotalCents
		}
		fmt.Fprintf(&sb, "%s %s: `%s%s` (%d tx)\n",
			emoji, s.CategoryName, sign, money.FormatCents(s.TotalCents), s.TxCount)
	}

	sb.WriteString("\n───────────\n")
	fmt.Fprintf(&sb, "Income:  `+%s`\nExpense: `-%s`\nNet:     `%s`",
		money.FormatCents(totalIncome),
		money.FormatCents(totalExpense),
		formatNet(totalIncome-totalExpense),
	)

	return sb.String()
}
