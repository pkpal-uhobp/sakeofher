package gateway

import (
	"context"

	"sakeofher/internal/domain"
)

type RemnawaveGateway interface {
	CreateUser(ctx context.Context, req domain.CreateRemnaUserRequest) (*domain.RemnaUser, error)
	EnableUser(ctx context.Context, remnaUUID string) error
	DisableUser(ctx context.Context, remnaUUID string) error
	DeleteUser(ctx context.Context, remnaUUID string) error
	ResetTraffic(ctx context.Context, remnaUUID string) error
	GetUserTraffic(ctx context.Context, remnaUUID string) (*domain.RemnaTraffic, error)
}

type TributeGateway interface{}

type CryptoBotGateway interface{}

type TelegramGateway interface {
	SendMessage(ctx context.Context, telegramID int64, text string) error
}

type Gateways struct {
	Remnawave RemnawaveGateway
	Tribute   TributeGateway
	CryptoBot CryptoBotGateway
	Telegram  TelegramGateway
}
