package domain

import "time"

type User struct {
	ID                int64       `json:"id"`
	TelegramID        int64       `json:"telegram_id"`
	TelegramUsername  *string     `json:"telegram_username,omitempty"`
	TelegramFirstName *string     `json:"telegram_first_name,omitempty"`
	TelegramLastName  *string     `json:"telegram_last_name,omitempty"`
	LanguageCode      *string     `json:"language_code,omitempty"`
	Alias             *string     `json:"alias,omitempty"`
	RemnaUUID         *string     `json:"remna_uuid,omitempty"`
	RemnaUsername     *string     `json:"remna_username,omitempty"`
	SubscriptionURL   *string     `json:"subscription_url,omitempty"`
	Status            UserStatus  `json:"status"`
	RemnaStatus       RemnaStatus `json:"remna_status"`
	DisabledAt        *time.Time  `json:"disabled_at,omitempty"`
	DeleteAfter       *time.Time  `json:"delete_after,omitempty"`
	DeletedAt         *time.Time  `json:"deleted_at,omitempty"`
	LastSeenAt        *time.Time  `json:"last_seen_at,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
}

type TelegramUserInput struct {
	TelegramID        int64   `json:"telegram_id"`
	TelegramUsername  *string `json:"telegram_username,omitempty"`
	TelegramFirstName *string `json:"telegram_first_name,omitempty"`
	TelegramLastName  *string `json:"telegram_last_name,omitempty"`
	LanguageCode      *string `json:"language_code,omitempty"`
}

type RemnaUserData struct {
	UUID            string      `json:"uuid"`
	Username        string      `json:"username"`
	SubscriptionURL string      `json:"subscription_url"`
	Status          RemnaStatus `json:"status"`
}
