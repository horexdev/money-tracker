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
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
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

// SendBudgetAlert sends a formatted budget alert to the user.
func (n *TelegramNotifier) SendBudgetAlert(ctx context.Context, chatID int64, categoryName string, spentPercent int, limitCents, spentCents int64) error {
	text := fmt.Sprintf(
		"⚠️ Budget Alert\n\nCategory: %s\nSpent: %d%% of limit\nUsed: %d / %d cents",
		categoryName, spentPercent, spentCents, limitCents,
	)
	return n.SendMessage(ctx, chatID, text)
}
