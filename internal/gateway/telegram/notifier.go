package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Notifier struct {
	token string
	log   *zap.Logger
	http  *http.Client
}

func NewNotifier(token string, log *zap.Logger) *Notifier {
	return &Notifier{token: token, log: log, http: &http.Client{Timeout: 10 * time.Second}}
}

func (n *Notifier) SendMessage(ctx context.Context, telegramID int64, text string) error {
	if n.token == "" {
		n.log.Info("telegram send message stub", zap.Int64("telegram_id", telegramID), zap.String("text", text))
		return nil
	}

	payload, _ := json.Marshal(map[string]any{
		"chat_id": telegramID,
		"text":    text,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.http.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send message: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram send message status: %d", resp.StatusCode)
	}
	return nil
}
