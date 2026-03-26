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
func StatsStartHandler(store *fsm.Store) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		userID := update.Message.From.ID
		store.SetState(ctx, userID, fsm.StateStatsWaitPeriod)

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "📊 Choose a period:",
			ReplyMarkup: buildPeriodKeyboard(),
		})
	}
}

// StatsPeriodHandler handles the period callback and renders the stats report.
func StatsPeriodHandler(store *fsm.Store, statsSvc *service.StatsService, log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		userID := query.From.ID

		b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{CallbackQueryID: query.ID})

		period := parsePeriodCallback(query.Data)
		from, to, err := service.PeriodRange(period)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: query.Message.Message.Chat.ID,
				Text:   "❌ Unknown period.",
			})
			store.Clear(ctx, userID)
			return
		}

		stats, err := statsSvc.ByCategory(ctx, userID, from, to)
		if err != nil {
			log.ErrorContext(ctx, "stats query failed", slog.String("error", err.Error()))
			sendErrorCallback(ctx, b, query.Message.Message.Chat.ID)
			store.Clear(ctx, userID)
			return
		}

		store.Clear(ctx, userID)

		if len(stats) == 0 {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: query.Message.Message.Chat.ID,
				Text:   "No transactions found for this period.",
			})
			return
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    query.Message.Message.Chat.ID,
			Text:      formatStats(stats, period),
			ParseMode: models.ParseModeMarkdown,
		})
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
