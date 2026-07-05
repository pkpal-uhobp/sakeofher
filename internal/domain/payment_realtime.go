package domain

import (
	"encoding/json"
	"time"
)

type TelegramStarsPaidInput struct {
	PaymentID  int64           `json:"payment_id"`
	ChargeID   string          `json:"charge_id"`
	EventID    string          `json:"event_id"`
	PaidAt     time.Time       `json:"paid_at"`
	RawPayload json.RawMessage `json:"raw_payload"`
}

type CreateCryptoBotPaymentInput struct {
	TelegramID     int64 `json:"telegram_id"`
	TariffPriceID int64 `json:"tariff_price_id"`
}
