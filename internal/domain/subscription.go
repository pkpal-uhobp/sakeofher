package domain

import "time"

type Subscription struct {
	ID                 int64              `json:"id"`
	UserID             int64              `json:"user_id"`
	LastPaymentID      *int64             `json:"last_payment_id,omitempty"`
	TariffID           *int64             `json:"tariff_id,omitempty"`
	Status             SubscriptionStatus `json:"status"`
	StartedAt          time.Time          `json:"started_at"`
	ExpiresAt          time.Time          `json:"expires_at"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	TrafficLimitBytes  int64              `json:"traffic_limit_bytes"`
	TrafficUsedBytes   int64              `json:"traffic_used_bytes"`
	PeriodStatus       PeriodStatus       `json:"period_status"`
	PublicToken         string             `json:"public_token"`

	LastRemnaCheckAt          *time.Time `json:"last_remna_check_at,omitempty"`
	LastExpireNotificationAt  *time.Time `json:"last_expire_notification_at,omitempty"`
	LastTrafficNotificationAt *time.Time `json:"last_traffic_notification_at,omitempty"`

	Notified3Days            bool `json:"notified_3_days"`
	Notified1Day             bool `json:"notified_1_day"`
	NotifiedExpired          bool `json:"notified_expired"`
	Traffic80Notified        bool `json:"traffic_80_notified"`
	Traffic95Notified        bool `json:"traffic_95_notified"`
	TrafficExhaustedNotified bool `json:"traffic_exhausted_notified"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SubscriptionWithUser struct {
	Subscription Subscription `json:"subscription"`
	User         User         `json:"user"`

	// Filled by repository methods that join tariffs.
	Tariff Tariff `json:"tariff"`
}

type PublicSubscription struct {
	Subscription    Subscription `json:"subscription"`
	User            User         `json:"user"`
	Tariff          Tariff       `json:"tariff"`
	SubscriptionURL *string      `json:"subscription_url,omitempty"`

	TelegramBotURL *string `json:"telegram_bot_url,omitempty"`
	BotURL         *string `json:"bot_url,omitempty"`
}

const BytesInGiB int64 = 1024 * 1024 * 1024

type SitePurchaseInput struct {
	TelegramID        int64   `json:"telegram_id"`
	TelegramUsername  *string `json:"telegram_username,omitempty"`
	TelegramFirstName *string `json:"telegram_first_name,omitempty"`
	TelegramLastName  *string `json:"telegram_last_name,omitempty"`
	LanguageCode      *string `json:"language_code,omitempty"`
	Alias             *string `json:"alias,omitempty"`
	TariffID          int64   `json:"tariff_id"`
	TrafficLimitGB    int64   `json:"traffic_limit_gb"`
}

type SiteRenewInput struct {
	TelegramID   int64  `json:"telegram_id"`
	PublicToken string `json:"public_token,omitempty"`
	TariffID    *int64 `json:"tariff_id,omitempty"`
}

func TrafficGBToBytes(gb int64) int64 {
	return gb * BytesInGiB
}

func TrafficBytesToGB(bytes int64) int64 {
	if bytes <= 0 {
		return 0
	}

	return bytes / BytesInGiB
}
