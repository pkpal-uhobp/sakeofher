package service

import (
	"context"
	"sakeofher/internal/gateway"
)

type NotificationService struct{ telegram gateway.TelegramGateway }

func NewNotificationService(telegram gateway.TelegramGateway) *NotificationService {
	return &NotificationService{telegram: telegram}
}

func (s *NotificationService) Send(ctx context.Context, telegramID int64, text string) error {
	if s.telegram == nil {
		return nil
	}
	return s.telegram.SendMessage(ctx, telegramID, text)
}
