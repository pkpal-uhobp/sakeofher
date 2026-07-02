package repository

import (
	"sakeofher/internal/repository/pool"
	"sakeofher/internal/repository/tx"
)

type Repositories struct {
	Tx *tx.Manager

	Users               *UserRepository
	Tariffs             *TariffRepository
	TariffPrices        *TariffPriceRepository
	Payments            *PaymentRepository
	PaymentEvents       *PaymentEventRepository
	Subscriptions       *SubscriptionRepository
	Admins              *AdminRepository
	Broadcasts          *BroadcastRepository
	BroadcastRecipients *BroadcastRecipientRepository
	RemnaSync           *RemnaSyncRepository
}

func NewRepositories(db *pool.ConnectionPool) *Repositories {
	txManager := tx.NewManager(db)
	return &Repositories{
		Tx:                  txManager,
		Users:               NewUserRepository(txManager),
		Tariffs:             NewTariffRepository(txManager),
		TariffPrices:        NewTariffPriceRepository(txManager),
		Payments:            NewPaymentRepository(txManager),
		PaymentEvents:       NewPaymentEventRepository(txManager),
		Subscriptions:       NewSubscriptionRepository(txManager),
		Admins:              NewAdminRepository(txManager),
		Broadcasts:          NewBroadcastRepository(txManager),
		BroadcastRecipients: NewBroadcastRecipientRepository(txManager),
		RemnaSync:           NewRemnaSyncRepository(txManager),
	}
}
