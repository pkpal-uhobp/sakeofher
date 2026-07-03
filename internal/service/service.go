package service

import (
	"context"
	"time"

	"sakeofher/internal/domain"
	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type UserService interface {
	GetOrCreateTelegramUser(ctx context.Context, input domain.TelegramUserInput) (*domain.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error)
}

type TariffService interface {
	ListActive(ctx context.Context) ([]domain.Tariff, error)
	ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error)
}

type AuthService interface {
	StartTelegramOAuth(ctx context.Context) (*domain.TelegramOAuthStart, string, string, string, error)
	FinishTelegramOAuth(ctx context.Context, input domain.TelegramOAuthCallbackInput) (*domain.AuthSession, error)
}

type SiteService interface {
	GetConfig(ctx context.Context) (*domain.SiteConfig, error)
	CreatePurchaseLink(ctx context.Context, input domain.SitePurchaseLinkInput) (*domain.SiteCheckoutLink, error)
	CreateRenewLink(ctx context.Context, input domain.SiteRenewLinkInput) (*domain.SiteCheckoutLink, error)
}

type PaymentService interface {
	CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error)
	MarkPaidForDev(ctx context.Context, paymentID int64, providerPaymentID string) (*domain.Payment, error)
	HandlePaymentPaid(ctx context.Context, input domain.PaymentPaidInput) error
	RetryFailedActivations(ctx context.Context, limit int) error
}

type SubscriptionService interface {
	GetPublicByToken(ctx context.Context, token string) (*domain.PublicSubscription, error)
	GetActiveByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error)
	GetLatestByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error)
	ActivateAfterPayment(ctx context.Context, paymentID int64) error
	DisableExpiredSubscriptions(ctx context.Context, limit int) error
	DeleteOldDisabledUsers(ctx context.Context, limit int) error
}

type NotificationService interface {
	Send(ctx context.Context, telegramID int64, text string) error
}

type AdminService interface{}

type BroadcastService interface{}

type WorkerService interface {
	ExpireSubscriptions(ctx context.Context) error
	DeleteOldDisabledUsers(ctx context.Context) error
	RetryFailedActivations(ctx context.Context) error
}

type Services struct {
	Auth          AuthService
	Users         UserService
	Tariffs       TariffService
	Site          SiteService
	Payments      PaymentService
	Subscriptions SubscriptionService
	Admins        AdminService
	Broadcasts    BroadcastService
	Notifications NotificationService
	Workers       WorkerService
}

func NewServices(repo *repository.Repositories, gates gateway.Gateways, telegramBotUsername string, publicURL string, subscriptionPathSecret string, jwtSecret string, jwtAccessTTL time.Duration, telegramOAuthRedirectURI string) *Services {
	notifications := NewNotificationService(gates.Telegram)
	subscriptions := NewSubscriptionService(repo, gates.Remnawave, notifications)
	payments := NewPaymentService(repo, gates, subscriptions)
	workers := NewWorkerService(subscriptions, payments)

	return &Services{
		Auth:          NewAuthService(repo, gates.TelegramOAuth, jwtSecret, jwtAccessTTL, telegramOAuthRedirectURI),
		Users:         NewUserService(repo),
		Tariffs:       NewTariffService(repo),
		Site:          NewSiteService(repo, telegramBotUsername, publicURL, subscriptionPathSecret),
		Payments:      payments,
		Subscriptions: subscriptions,
		Admins:        NewAdminService(repo),
		Broadcasts:    NewBroadcastService(repo, notifications),
		Notifications: notifications,
		Workers:       workers,
	}
}
