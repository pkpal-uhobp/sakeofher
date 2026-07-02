package service

import (
	"context"

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

type PaymentService interface {
	CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error)
	MarkPaidForDev(ctx context.Context, paymentID int64, providerPaymentID string) (*domain.Payment, error)
	HandlePaymentPaid(ctx context.Context, input domain.PaymentPaidInput) error
	RetryFailedActivations(ctx context.Context, limit int) error
}

type SubscriptionService interface {
	GetPublicByToken(ctx context.Context, token string) (*domain.PublicSubscription, error)
	GetActiveByTelegramID(ctx context.Context, telegramID int64) (*domain.PublicSubscription, error)
	ActivateAfterPayment(ctx context.Context, paymentID int64) error
	DisableExpiredSubscriptions(ctx context.Context, limit int) error
	DeleteOldDisabledUsers(ctx context.Context, limit int) error
	PurchaseFromSite(ctx context.Context, input domain.SitePurchaseInput) (*domain.PublicSubscription, error)
	RenewFromSite(ctx context.Context, input domain.SiteRenewInput) (*domain.PublicSubscription, error)
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
	Users         UserService
	Tariffs       TariffService
	Payments      PaymentService
	Subscriptions SubscriptionService
	Admins        AdminService
	Broadcasts    BroadcastService
	Notifications NotificationService
	Workers       WorkerService
}

func NewServices(repo *repository.Repositories, gates gateway.Gateways) *Services {
	notifications := NewNotificationService(gates.Telegram)
	subscriptions := NewSubscriptionService(repo, gates.Remnawave, notifications)
	payments := NewPaymentService(repo, gates, subscriptions)
	workers := NewWorkerService(subscriptions, payments)

	return &Services{
		Users:         NewUserService(repo),
		Tariffs:       NewTariffService(repo),
		Payments:      payments,
		Subscriptions: subscriptions,
		Admins:        NewAdminService(repo),
		Broadcasts:    NewBroadcastService(repo, notifications),
		Notifications: notifications,
		Workers:       workers,
	}
}
