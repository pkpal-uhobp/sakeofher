package domain

import "time"

type Payment struct {
	ID                int64
	UserID            int64
	TariffID          int64
	TariffPriceID     int64
	Provider          PaymentProvider
	PaymentMethod     PaymentMethod
	Status            PaymentStatus
	Currency          string
	AmountMinor       *int64
	StarsAmount       *int64
	ProviderPaymentID *string
	PaymentURL        *string
	PaidAsset         *string
	PaidAmount        *string
	FeeAsset          *string
	FeeAmount         *string
	ExpiresAt         *time.Time
	PaidAt            *time.Time
	RawPayload        []byte
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PaymentEvent struct {
	ID          int64
	Provider    PaymentProvider
	EventID     string
	PaymentID   *int64
	EventType   string
	RawPayload  []byte
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

type PaymentPaidInput struct {
	Provider          PaymentProvider
	EventID           string
	EventType         string
	ProviderPaymentID string
	PaidAt            time.Time
	RawPayload        []byte
}
