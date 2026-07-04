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

	List(ctx context.Context, input domain.UserListInput) (*domain.UserListResponse, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	Update(ctx context.Context, id int64, input domain.UpdateUserInput) (*domain.User, error)
	Block(ctx context.Context, id int64) (*domain.User, error)
	Unblock(ctx context.Context, id int64) (*domain.User, error)
	MarkDeleted(ctx context.Context, id int64) (*domain.User, error)
}

type TariffService interface {
	ListActive(ctx context.Context) ([]domain.Tariff, error)
	ListActiveWithPrices(ctx context.Context) ([]domain.TariffWithPrices, error)

	ListAll(ctx context.Context) ([]domain.Tariff, error)
	GetByID(ctx context.Context, id int64) (*domain.Tariff, error)
	Create(ctx context.Context, input domain.CreateTariffInput) (*domain.Tariff, error)
	Update(ctx context.Context, id int64, input domain.UpdateTariffInput) (*domain.Tariff, error)
	Enable(ctx context.Context, id int64) (*domain.Tariff, error)
	Disable(ctx context.Context, id int64) (*domain.Tariff, error)
}

type AuthService interface {
	Login(ctx context.Context, input domain.LoginInput) (*domain.AuthSession, error)
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
	SyncRemnaUsage(ctx context.Context, limit int) error
	ResetTrafficPeriods(ctx context.Context, limit int) error

	List(ctx context.Context, input domain.SubscriptionListInput) (*domain.SubscriptionListResponse, error)
	GetByID(ctx context.Context, id int64) (*domain.PublicSubscription, error)
	CreateManual(ctx context.Context, input domain.CreateManualSubscriptionInput) (*domain.PublicSubscription, error)
	Extend(ctx context.Context, id int64, input domain.ExtendSubscriptionInput) (*domain.PublicSubscription, error)
	Update(ctx context.Context, id int64, input domain.UpdateSubscriptionInput) (*domain.PublicSubscription, error)
	UpdateTrafficLimit(ctx context.Context, id int64, input domain.UpdateTrafficLimitInput) (*domain.PublicSubscription, error)
	Disable(ctx context.Context, id int64) (*domain.PublicSubscription, error)
	Enable(ctx context.Context, id int64) (*domain.PublicSubscription, error)
	Cancel(ctx context.Context, id int64) (*domain.PublicSubscription, error)
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
	SyncUsage(ctx context.Context) error
	ResetTrafficPeriods(ctx context.Context) error
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

func NewServices(
	repo *repository.Repositories,
	gates gateway.Gateways,
	telegramBotUsername string,
	publicURL string,
	subscriptionPathSecret string,
	adminUsername string,
	adminPassword string,
	jwtSecret string,
	jwtAccessTTL time.Duration,
) *Services {
	notifications := NewNotificationService(gates.Telegram)
	subscriptions := NewSubscriptionService(repo, gates.Remnawave, notifications)
	payments := NewPaymentService(repo, gates, subscriptions)
	workers := NewWorkerService(subscriptions, payments)

	return &Services{
		Auth:          NewAuthService(adminUsername, adminPassword, jwtSecret, jwtAccessTTL),
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
