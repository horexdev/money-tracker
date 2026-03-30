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

// categoryNames maps English category keys to their localized names.
var categoryNames = map[string]map[string]string{
	"en": {"Food": "Food", "Transport": "Transport", "Entertainment": "Entertainment", "Shopping": "Shopping", "Health": "Health", "Salary": "Salary", "Freelance": "Freelance", "Rent": "Rent", "Coffee": "Coffee", "Bills": "Bills", "Education": "Education", "Travel": "Travel", "Gifts": "Gifts", "Sport": "Sport", "Beauty": "Beauty", "Pets": "Pets", "Business": "Business", "Housing": "Housing", "Investments": "Investments", "Other": "Other", "Savings": "Savings", "Transfer": "Transfer"},
	"ru": {"Food": "Еда", "Transport": "Транспорт", "Entertainment": "Развлечения", "Shopping": "Покупки", "Health": "Здоровье", "Salary": "Зарплата", "Freelance": "Фриланс", "Rent": "Аренда", "Coffee": "Кофе", "Bills": "Счета", "Education": "Образование", "Travel": "Путешествия", "Gifts": "Подарки", "Sport": "Спорт", "Beauty": "Красота", "Pets": "Питомцы", "Business": "Бизнес", "Housing": "Жильё", "Investments": "Инвестиции", "Other": "Прочее", "Savings": "Накопления", "Transfer": "Перевод"},
	"uk": {"Food": "Їжа", "Transport": "Транспорт", "Entertainment": "Розваги", "Shopping": "Покупки", "Health": "Здоров'я", "Salary": "Зарплата", "Freelance": "Фриланс", "Rent": "Оренда", "Coffee": "Кава", "Bills": "Рахунки", "Education": "Освіта", "Travel": "Подорожі", "Gifts": "Подарунки", "Sport": "Спорт", "Beauty": "Краса", "Pets": "Домашні тварини", "Business": "Бізнес", "Housing": "Житло", "Investments": "Інвестиції", "Other": "Інше", "Savings": "Накопичення", "Transfer": "Переказ"},
	"be": {"Food": "Ежа", "Transport": "Транспарт", "Entertainment": "Забавы", "Shopping": "Пакупкі", "Health": "Здароўе", "Salary": "Зарплата", "Freelance": "Фрыланс", "Rent": "Арэнда", "Coffee": "Кава", "Bills": "Рахункі", "Education": "Адукацыя", "Travel": "Падарожжы", "Gifts": "Падарункі", "Sport": "Спорт", "Beauty": "Краса", "Pets": "Хатнія жывёлы", "Business": "Бізнес", "Housing": "Жыллё", "Investments": "Інвестыцыі", "Other": "Іншае", "Savings": "Назапашванні", "Transfer": "Пераклад"},
	"kk": {"Food": "Тамақ", "Transport": "Көлік", "Entertainment": "Ойын-сауық", "Shopping": "Сауда", "Health": "Денсаулық", "Salary": "Жалақы", "Freelance": "Фриланс", "Rent": "Жалдау", "Coffee": "Кофе", "Bills": "Шоттар", "Education": "Білім", "Travel": "Саяхат", "Gifts": "Сыйлықтар", "Sport": "Спорт", "Beauty": "Сұлулық", "Pets": "Үй жануарлары", "Business": "Бизнес", "Housing": "Тұрғын үй", "Investments": "Инвестициялар", "Other": "Басқа", "Savings": "Жинақтар", "Transfer": "Аудару"},
	"uz": {"Food": "Oziq-ovqat", "Transport": "Transport", "Entertainment": "Ko'ngil ochar", "Shopping": "Xarid", "Health": "Salomatlik", "Salary": "Maosh", "Freelance": "Frilanser", "Rent": "Ijara", "Coffee": "Qahva", "Bills": "To'lovlar", "Education": "Ta'lim", "Travel": "Sayohat", "Gifts": "Sovg'alar", "Sport": "Sport", "Beauty": "Go'zallik", "Pets": "Uy hayvonlari", "Business": "Biznes", "Housing": "Uy-joy", "Investments": "Investitsiyalar", "Other": "Boshqa", "Savings": "Jamg'armalar", "Transfer": "O'tkazma"},
	"es": {"Food": "Comida", "Transport": "Transporte", "Entertainment": "Entretenimiento", "Shopping": "Compras", "Health": "Salud", "Salary": "Salario", "Freelance": "Freelance", "Rent": "Alquiler", "Coffee": "Café", "Bills": "Facturas", "Education": "Educación", "Travel": "Viajes", "Gifts": "Regalos", "Sport": "Deporte", "Beauty": "Belleza", "Pets": "Mascotas", "Business": "Negocios", "Housing": "Vivienda", "Investments": "Inversiones", "Other": "Otros", "Savings": "Ahorros", "Transfer": "Transferencia"},
	"de": {"Food": "Essen", "Transport": "Transport", "Entertainment": "Unterhaltung", "Shopping": "Einkaufen", "Health": "Gesundheit", "Salary": "Gehalt", "Freelance": "Freelance", "Rent": "Miete", "Coffee": "Kaffee", "Bills": "Rechnungen", "Education": "Bildung", "Travel": "Reisen", "Gifts": "Geschenke", "Sport": "Sport", "Beauty": "Schönheit", "Pets": "Haustiere", "Business": "Geschäft", "Housing": "Wohnen", "Investments": "Investitionen", "Other": "Sonstiges", "Savings": "Ersparnisse", "Transfer": "Überweisung"},
	"it": {"Food": "Cibo", "Transport": "Trasporti", "Entertainment": "Intrattenimento", "Shopping": "Shopping", "Health": "Salute", "Salary": "Stipendio", "Freelance": "Freelance", "Rent": "Affitto", "Coffee": "Caffè", "Bills": "Bollette", "Education": "Istruzione", "Travel": "Viaggi", "Gifts": "Regali", "Sport": "Sport", "Beauty": "Bellezza", "Pets": "Animali", "Business": "Affari", "Housing": "Abitazione", "Investments": "Investimenti", "Other": "Altro", "Savings": "Risparmi", "Transfer": "Bonifico"},
	"fr": {"Food": "Alimentation", "Transport": "Transport", "Entertainment": "Divertissement", "Shopping": "Shopping", "Health": "Santé", "Salary": "Salaire", "Freelance": "Freelance", "Rent": "Loyer", "Coffee": "Café", "Bills": "Factures", "Education": "Éducation", "Travel": "Voyages", "Gifts": "Cadeaux", "Sport": "Sport", "Beauty": "Beauté", "Pets": "Animaux", "Business": "Affaires", "Housing": "Logement", "Investments": "Investissements", "Other": "Autre", "Savings": "Épargne", "Transfer": "Virement"},
	"pt": {"Food": "Alimentação", "Transport": "Transporte", "Entertainment": "Entretenimento", "Shopping": "Compras", "Health": "Saúde", "Salary": "Salário", "Freelance": "Freelance", "Rent": "Aluguel", "Coffee": "Café", "Bills": "Contas", "Education": "Educação", "Travel": "Viagens", "Gifts": "Presentes", "Sport": "Esporte", "Beauty": "Beleza", "Pets": "Animais de estimação", "Business": "Negócios", "Housing": "Moradia", "Investments": "Investimentos", "Other": "Outros", "Savings": "Poupanças", "Transfer": "Transferência"},
	"nl": {"Food": "Eten", "Transport": "Vervoer", "Entertainment": "Vermaak", "Shopping": "Winkelen", "Health": "Gezondheid", "Salary": "Salaris", "Freelance": "Freelance", "Rent": "Huur", "Coffee": "Koffie", "Bills": "Rekeningen", "Education": "Onderwijs", "Travel": "Reizen", "Gifts": "Cadeaus", "Sport": "Sport", "Beauty": "Schoonheid", "Pets": "Huisdieren", "Business": "Zakelijk", "Housing": "Wonen", "Investments": "Investeringen", "Other": "Overig", "Savings": "Spaargeld", "Transfer": "Overboeking"},
	"ar": {"Food": "طعام", "Transport": "مواصلات", "Entertainment": "ترفيه", "Shopping": "تسوق", "Health": "صحة", "Salary": "راتب", "Freelance": "عمل حر", "Rent": "إيجار", "Coffee": "قهوة", "Bills": "فواتير", "Education": "تعليم", "Travel": "سفر", "Gifts": "هدايا", "Sport": "رياضة", "Beauty": "جمال", "Pets": "حيوانات أليفة", "Business": "أعمال", "Housing": "سكن", "Investments": "استثمارات", "Other": "أخرى", "Savings": "المدخرات", "Transfer": "تحويل"},
	"tr": {"Food": "Yemek", "Transport": "Ulaşım", "Entertainment": "Eğlence", "Shopping": "Alışveriş", "Health": "Sağlık", "Salary": "Maaş", "Freelance": "Serbest çalışma", "Rent": "Kira", "Coffee": "Kahve", "Bills": "Faturalar", "Education": "Eğitim", "Travel": "Seyahat", "Gifts": "Hediyeler", "Sport": "Spor", "Beauty": "Güzellik", "Pets": "Evcil hayvanlar", "Business": "İş", "Housing": "Konut", "Investments": "Yatırımlar", "Other": "Diğer", "Savings": "Tasarruflar", "Transfer": "Transfer"},
	"ko": {"Food": "음식", "Transport": "교통", "Entertainment": "오락", "Shopping": "쇼핑", "Health": "건강", "Salary": "급여", "Freelance": "프리랜서", "Rent": "임대료", "Coffee": "커피", "Bills": "청구서", "Education": "교육", "Travel": "여행", "Gifts": "선물", "Sport": "스포츠", "Beauty": "미용", "Pets": "반려동물", "Business": "비즈니스", "Housing": "주거", "Investments": "투자", "Other": "기타", "Savings": "저축", "Transfer": "이체"},
	"ms": {"Food": "Makanan", "Transport": "Pengangkutan", "Entertainment": "Hiburan", "Shopping": "Membeli-belah", "Health": "Kesihatan", "Salary": "Gaji", "Freelance": "Bebas", "Rent": "Sewa", "Coffee": "Kopi", "Bills": "Bil", "Education": "Pendidikan", "Travel": "Pelancongan", "Gifts": "Hadiah", "Sport": "Sukan", "Beauty": "Kecantikan", "Pets": "Haiwan peliharaan", "Business": "Perniagaan", "Housing": "Perumahan", "Investments": "Pelaburan", "Other": "Lain-lain", "Savings": "Simpanan", "Transfer": "Pindahan"},
	"id": {"Food": "Makanan", "Transport": "Transportasi", "Entertainment": "Hiburan", "Shopping": "Belanja", "Health": "Kesehatan", "Salary": "Gaji", "Freelance": "Lepas", "Rent": "Sewa", "Coffee": "Kopi", "Bills": "Tagihan", "Education": "Pendidikan", "Travel": "Perjalanan", "Gifts": "Hadiah", "Sport": "Olahraga", "Beauty": "Kecantikan", "Pets": "Hewan peliharaan", "Business": "Bisnis", "Housing": "Perumahan", "Investments": "Investasi", "Other": "Lainnya", "Savings": "Tabungan", "Transfer": "Transfer"},
}

// translateCategory returns the localized category name, falling back to the English key.
func translateCategory(lang, englishName string) string {
	if names, ok := categoryNames[lang]; ok {
		if translated, ok := names[englishName]; ok {
			return translated
		}
	}
	return englishName
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

	localizedCategory := translateCategory(lang, categoryName)
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
		labels[1], localizedCategory,
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
