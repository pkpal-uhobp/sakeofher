package service

import (
	"context"
	"sakeofher/internal/gateway"
)

type notificationService struct{ telegram gateway.TelegramGateway }

func NewNotificationService(telegram gateway.TelegramGateway) NotificationService {
	return &notificationService{telegram: telegram}
}

func (s *notificationService) Send(ctx context.Context, telegramID int64, text string) error {
	if s.telegram == nil {
		return nil
	}
	return s.telegram.SendMessage(ctx, telegramID, text)
}
