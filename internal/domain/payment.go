package domain

import (
	"encoding/json"
	"time"
)

type Payment struct {
	ID                int64           `json:"id"`
	UserID            int64           `json:"user_id"`
	TariffID          int64           `json:"tariff_id"`
	TariffPriceID     *int64          `json:"tariff_price_id,omitempty"`
	Provider          PaymentProvider `json:"provider"`
	PaymentMethod     PaymentMethod   `json:"payment_method"`
	Currency          string          `json:"currency"`
	AmountMinor       *int64          `json:"amount_minor,omitempty"`
	StarsAmount       *int64          `json:"stars_amount,omitempty"`
	DurationDays      int             `json:"duration_days"`
	PeriodDays        int             `json:"period_days"`
	TrafficLimitBytes int64           `json:"traffic_limit_bytes"`
	Status            PaymentStatus   `json:"status"`
	ProviderPaymentID *string         `json:"provider_payment_id,omitempty"`
	PaymentURL        *string         `json:"payment_url,omitempty"`
	PaidAsset         *string         `json:"paid_asset,omitempty"`
	PaidAmount        *string         `json:"paid_amount,omitempty"`
	FeeAsset          *string         `json:"fee_asset,omitempty"`
	FeeAmount         *string         `json:"fee_amount,omitempty"`
	ExpiresAt         *time.Time      `json:"expires_at,omitempty"`
	PaidAt            *time.Time      `json:"paid_at,omitempty"`
	ActivatedAt       *time.Time      `json:"activated_at,omitempty"`
	RawPayload        json.RawMessage `json:"raw_payload,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

type PaymentEvent struct {
	ID          int64           `json:"id"`
	Provider    PaymentProvider `json:"provider"`
	EventID     string          `json:"event_id"`
	PaymentID   *int64          `json:"payment_id,omitempty"`
	EventType   string          `json:"event_type"`
	RawPayload  json.RawMessage `json:"raw_payload"`
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type CreatePaymentInput struct {
	TelegramID    int64 `json:"telegram_id"`
	TariffPriceID int64 `json:"tariff_price_id"`
}

type PaymentPaidInput struct {
	Provider          PaymentProvider `json:"provider"`
	EventID           string          `json:"event_id"`
	EventType         string          `json:"event_type"`
	ProviderPaymentID string          `json:"provider_payment_id"`
	PaidAt            time.Time       `json:"paid_at"`
	RawPayload        json.RawMessage `json:"raw_payload"`
}
