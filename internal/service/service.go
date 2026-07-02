package service

import (
	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type Services struct {
	Users         *UserService
	Tariffs       *TariffService
	Payments      *PaymentService
	Subscriptions *SubscriptionService
	Admins        *AdminService
	Broadcasts    *BroadcastService
	Notifications *NotificationService
	Workers       *WorkerService
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
