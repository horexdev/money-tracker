package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// TelegramNotifier sends messages to users via the Telegram Bot API.
type TelegramNotifier struct {
	token  string
	client *http.Client
	log    *slog.Logger
}

// NewTelegramNotifier creates a notifier using the given bot token.
func NewTelegramNotifier(token string, log *slog.Logger) *TelegramNotifier {
	return &TelegramNotifier{
		token: token,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

type sendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// SendMessage sends a plain text message to a Telegram chat.
func (n *TelegramNotifier) SendMessage(ctx context.Context, chatID int64, text string) error {
	body, err := json.Marshal(sendMessageRequest{ChatID: chatID, Text: text})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	n.log.DebugContext(ctx, "telegram message sent", slog.Int64("chat_id", chatID))
	return nil
}

// budgetAlertTexts holds localized strings for budget alerts.
// Keys are language codes matching domain.Language constants.
var budgetAlertTexts = map[string][4]string{
	// [0]=title, [1]=category label, [2]=spent label, [3]=remaining label
	"en": {"Budget Alert", "Category", "Spent", "Remaining"},
	"ru": {"Уведомление о бюджете", "Категория", "Потрачено", "Остаток"},
	"uk": {"Сповіщення про бюджет", "Категорія", "Витрачено", "Залишок"},
	"be": {"Апавяшчэнне аб бюджэце", "Катэгорыя", "Выдаткавана", "Застаецца"},
	"kk": {"Бюджет туралы хабарлама", "Санат", "Жұмсалды", "Қалды"},
	"uz": {"Byudjet haqida xabar", "Kategoriya", "Sarflandi", "Qoldi"},
	"es": {"Alerta de presupuesto", "Categoría", "Gastado", "Restante"},
	"de": {"Budget-Benachrichtigung", "Kategorie", "Ausgegeben", "Verbleibend"},
	"it": {"Avviso budget", "Categoria", "Speso", "Rimanente"},
	"fr": {"Alerte budget", "Catégorie", "Dépensé", "Restant"},
	"pt": {"Alerta de orçamento", "Categoria", "Gasto", "Restante"},
	"nl": {"Budget melding", "Categorie", "Uitgegeven", "Resterend"},
	"ar": {"تنبيه الميزانية", "الفئة", "المُنفق", "المتبقي"},
	"tr": {"Bütçe Uyarısı", "Kategori", "Harcandı", "Kalan"},
	"ko": {"예산 알림", "카테고리", "지출됨", "남음"},
	"ms": {"Amaran belanjawan", "Kategori", "Dibelanjakan", "Baki"},
	"id": {"Peringatan anggaran", "Kategori", "Terpakai", "Sisa"},
}

func alertEmoji(percent int) string {
	switch {
	case percent >= 100:
		return "🔴"
	case percent >= 95:
		return "🟠"
	case percent >= 75:
		return "🟡"
	default:
		return "🟢"
	}
}

func formatMoney(cents int64, currency string) string {
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	return fmt.Sprintf("%d.%02d %s", whole, frac, currency)
}

// SendBudgetAlert sends a formatted, localized budget alert to the user.
func (n *TelegramNotifier) SendBudgetAlert(ctx context.Context, chatID int64, lang, categoryName, currencyCode string, spentPercent int, limitCents, spentCents int64) error {
	labels, ok := budgetAlertTexts[lang]
	if !ok {
		labels = budgetAlertTexts["en"]
	}

	emoji := alertEmoji(spentPercent)
	remainingCents := limitCents - spentCents
	if remainingCents < 0 {
		remainingCents = 0
	}

	var progressBar string
	filled := spentPercent / 10
	if filled > 10 {
		filled = 10
	}
	for i := 0; i < 10; i++ {
		if i < filled {
			progressBar += "▓"
		} else {
			progressBar += "░"
		}
	}

	text := fmt.Sprintf(
		"%s *%s*\n\n%s: %s\n%s %d%%\n\n%s: %s\n%s: %s",
		emoji, labels[0],
		labels[1], categoryName,
		progressBar, spentPercent,
		labels[2], formatMoney(spentCents, currencyCode),
		labels[3], formatMoney(remainingCents, currencyCode),
	)

	body, err := json.Marshal(sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	})
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status %d", resp.StatusCode)
	}

	n.log.InfoContext(ctx, "budget alert sent",
		slog.Int64("chat_id", chatID),
		slog.String("category", categoryName),
		slog.Int("percent", spentPercent),
	)
	return nil
}
