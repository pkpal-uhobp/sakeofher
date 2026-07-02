package domain

import "time"

type User struct {
	ID               int64
	TelegramID       int64
	TelegramUsername *string
	FirstName        *string
	LastName         *string

	RemnaUUID       *string
	RemnaUsername   *string
	SubscriptionURL *string
	PublicToken     string
	RemnaStatus     RemnaStatus

	DisabledAt  *time.Time
	DeleteAfter *time.Time
	DeletedAt   *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

type TelegramUserInput struct {
	TelegramID       int64
	TelegramUsername *string
	FirstName        *string
	LastName         *string
}

type RemnaUserData struct {
	UUID            string
	Username        string
	SubscriptionURL string
	Status          RemnaStatus
}
