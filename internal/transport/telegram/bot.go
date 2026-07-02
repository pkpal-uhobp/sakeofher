package telegramtransport

import (
	"context"
	"go.uber.org/zap"
	"sakeofher/internal/service"
)

type Bot struct {
	token    string
	services *service.Services
	log      *zap.Logger
}

func NewBot(token string, services *service.Services, log *zap.Logger) *Bot {
	return &Bot{token: token, services: services, log: log}
}

func (b *Bot) Run(ctx context.Context) error {
	b.log.Info("telegram bot started stub")
	<-ctx.Done()
	return ctx.Err()
}
