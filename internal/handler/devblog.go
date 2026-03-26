package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/horexdev/money-tracker/devblogs"
)

const devblogPageSize = 5

// DevblogHandler handles the "📓 Devblog" reply keyboard button.
// Shows the first page of devblog entries.
func DevblogHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		if update.Message == nil {
			return
		}
		chatID := update.Message.Chat.ID

		text, kb := buildDevblogListPage(1, log)
		if _, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      chatID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "devblog: failed to send list", slog.String("error", err.Error()))
		}
	}
}

// DevblogCallbackHandler handles all "dlog:*" callback queries.
// Routes: "dlog:list:{page}" and "dlog:read:{filename}".
func DevblogCallbackHandler(log *slog.Logger) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		query := update.CallbackQuery
		if query == nil {
			return
		}

		if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: query.ID,
		}); err != nil {
			log.ErrorContext(ctx, "devblog: failed to answer callback", slog.String("error", err.Error()))
		}

		chatID := query.Message.Message.Chat.ID
		msgID := query.Message.Message.ID
		data := query.Data

		var text string
		var kb *models.InlineKeyboardMarkup

		switch {
		case strings.HasPrefix(data, "dlog:list:"):
			pageStr := strings.TrimPrefix(data, "dlog:list:")
			page, err := strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				page = 1
			}
			text, kb = buildDevblogListPage(page, log)

		case strings.HasPrefix(data, "dlog:read:"):
			filename := strings.TrimPrefix(data, "dlog:read:")
			text, kb = buildDevblogEntryView(filename, log)

		default:
			return
		}

		if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
			ChatID:      chatID,
			MessageID:   msgID,
			Text:        text,
			ParseMode:   models.ParseModeHTML,
			ReplyMarkup: kb,
		}); err != nil {
			log.ErrorContext(ctx, "devblog: failed to edit message", slog.String("error", err.Error()))
		}
	}
}

// buildDevblogListPage renders the paginated devblog entry list.
func buildDevblogListPage(page int, log *slog.Logger) (string, *models.InlineKeyboardMarkup) {
	entries, err := devblog.List()
	if err != nil {
		log.Error("devblog: failed to list entries", slog.String("error", err.Error()))
		return "Something went wrong loading the devblog. Please try again.", nil
	}

	if len(entries) == 0 {
		return "<b>📓 Devblog</b>\n\nNo release notes yet.", nil
	}

	totalPages := (len(entries) + devblogPageSize - 1) / devblogPageSize
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * devblogPageSize
	end := start + devblogPageSize
	if end > len(entries) {
		end = len(entries)
	}
	pageEntries := entries[start:end]

	var sb strings.Builder
	fmt.Fprintf(&sb, "<b>📓 Devblog</b>  <i>Page %d/%d</i>\n\n", page, totalPages)
	fmt.Fprintf(&sb, "Tap an entry to read the release notes.\n")

	// Build inline keyboard: one button per entry + pagination row.
	var rows [][]models.InlineKeyboardButton
	for _, e := range pageEntries {
		label := fmt.Sprintf("%s — %s", e.Version, e.Date.Format("02 Jan 2006"))
		rows = append(rows, []models.InlineKeyboardButton{
			{
				Text:         label,
				CallbackData: fmt.Sprintf("dlog:read:%s", e.Filename),
			},
		})
	}

	// Pagination row.
	paginationKb := buildDevblogListKeyboard(page, totalPages)
	rows = append(rows, paginationKb.InlineKeyboard[0])

	return sb.String(), &models.InlineKeyboardMarkup{InlineKeyboard: rows}
}

// buildDevblogEntryView renders a single devblog entry as Telegram HTML.
func buildDevblogEntryView(filename string, log *slog.Logger) (string, *models.InlineKeyboardMarkup) {
	raw, err := devblog.Content(filename)
	if err != nil {
		log.Error("devblog: failed to read entry", slog.String("filename", filename), slog.String("error", err.Error()))
		return "Could not load this devblog entry.", buildDevblogEntryKeyboard(1)
	}

	text := markdownToTelegramHTML(raw)

	const maxLen = 3800
	if len(text) > maxLen {
		text = text[:maxLen] + "\n\n<i>(truncated — see devblogs/ for full text)</i>"
	}

	return text, buildDevblogEntryKeyboard(1)
}

// markdownToTelegramHTML converts a subset of Markdown to Telegram HTML.
// Handles: headings (#, ##), bullet lists (-), inline code (`code`).
// All other text is HTML-escaped.
func markdownToTelegramHTML(raw []byte) string {
	lines := strings.Split(string(raw), "\n")
	var sb strings.Builder

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "# "):
			sb.WriteString("<b>")
			sb.WriteString(escapeHTML(strings.TrimPrefix(line, "# ")))
			sb.WriteString("</b>")
		case strings.HasPrefix(line, "## "):
			sb.WriteString("\n<b>")
			sb.WriteString(escapeHTML(strings.TrimPrefix(line, "## ")))
			sb.WriteString("</b>")
		case strings.HasPrefix(line, "### "):
			sb.WriteString("<b>")
			sb.WriteString(escapeHTML(strings.TrimPrefix(line, "### ")))
			sb.WriteString("</b>")
		default:
			sb.WriteString(convertInlineMarkdown(escapeHTML(line)))
		}
		sb.WriteByte('\n')
	}

	return strings.TrimSpace(sb.String())
}

// convertInlineMarkdown converts inline backtick code spans in an already HTML-escaped line.
// Example: "`foo`" → "<code>foo</code>".
// Note: the input has already been HTML-escaped, so backticks are safe to split on.
func convertInlineMarkdown(line string) string {
	parts := strings.Split(line, "`")
	if len(parts) < 3 {
		return line
	}
	var sb strings.Builder
	for i, part := range parts {
		if i%2 == 1 {
			sb.WriteString("<code>")
			sb.WriteString(part)
			sb.WriteString("</code>")
		} else {
			sb.WriteString(part)
		}
	}
	return sb.String()
}
