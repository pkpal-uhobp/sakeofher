package telegram

import (
	"context"
	"go.uber.org/zap"
)

type Notifier struct {
	token string
	log   *zap.Logger
}

func NewNotifier(token string, log *zap.Logger) *Notifier {
	return &Notifier{token: token, log: log}
}

func (n *Notifier) SendMessage(ctx context.Context, telegramID int64, text string) error {
	n.log.Info("telegram send message stub", zap.Int64("telegram_id", telegramID), zap.String("text", text))
	return nil
}
