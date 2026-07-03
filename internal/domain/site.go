package domain

import "time"

type SiteConfig struct {
	TelegramBotUsername    string `json:"telegram_bot_username"`
	TelegramBotURL         string `json:"telegram_bot_url"`
	PaymentsLocation       string `json:"payments_location"`
	PublicURL              string `json:"public_url"`
	SubscriptionPathSecret string `json:"subscription_path_secret"`
	SubscriptionURLPattern string `json:"subscription_url_pattern"`
}

type SiteCheckoutAction string

const (
	SiteCheckoutActionPurchase SiteCheckoutAction = "purchase"
	SiteCheckoutActionRenew    SiteCheckoutAction = "renew"
)

type SitePurchaseLinkInput struct {
	TariffID       int64 `json:"tariff_id"`
	TrafficLimitGB int64 `json:"traffic_limit_gb"`
}

type SiteRenewLinkInput struct {
	PublicToken string `json:"public_token"`
	TariffID    *int64 `json:"tariff_id,omitempty"`
}

type SiteCheckoutLink struct {
	Action               SiteCheckoutAction `json:"action"`
	StartPayload         string             `json:"start_payload"`
	TelegramBotURL       string             `json:"telegram_bot_url"`
	TelegramBotUsername  string             `json:"telegram_bot_username"`
	Tariff               Tariff             `json:"tariff"`
	TrafficLimitGB       int64              `json:"traffic_limit_gb"`
	TrafficLimitBytes    int64              `json:"traffic_limit_bytes"`
	PublicToken          string             `json:"public_token,omitempty"`
	CurrentExpiresAt     *time.Time         `json:"current_expires_at,omitempty"`
	NextExpiresAtPreview time.Time          `json:"next_expires_at_preview"`
	Note                 string             `json:"note"`
}
