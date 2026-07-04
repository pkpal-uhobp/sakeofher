package domain

import "time"

type Tariff struct {
	ID                int64     `json:"id"`
	Code              string    `json:"code"`
	Title             string    `json:"title"`
	Description       *string   `json:"description,omitempty"`
	DurationDays      int       `json:"duration_days"`
	PeriodDays        int       `json:"period_days"`
	TrafficLimitBytes int64     `json:"traffic_limit_bytes"`
	PriceRub          int64     `json:"price_rub"`
	IsActive          bool      `json:"is_active"`
	SortOrder         int       `json:"sort_order"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type TariffPrice struct {
	ID             int64           `json:"id"`
	TariffID       int64           `json:"tariff_id"`
	Provider       PaymentProvider `json:"provider"`
	PaymentMethod  PaymentMethod   `json:"payment_method"`
	Currency       string          `json:"currency"`
	AmountMinor    *int64          `json:"amount_minor,omitempty"`
	StarsAmount    *int64          `json:"stars_amount,omitempty"`
	AcceptedAssets []string        `json:"accepted_assets"`
	IsActive        bool            `json:"is_active"`
	SortOrder       int             `json:"sort_order"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type TariffWithPrices struct {
	Tariff
	Prices []TariffPrice `json:"prices"`
}

type TariffPriceWithTariff struct {
	Price  TariffPrice `json:"price"`
	Tariff Tariff      `json:"tariff"`
}
