package domain

import "time"

type Tariff struct {
	ID                int64
	Code              string
	Name              string
	Description       *string
	DurationDays      int
	TrafficLimitBytes int64
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TariffPrice struct {
	ID             int64
	TariffID       int64
	Provider       PaymentProvider
	PaymentMethod  PaymentMethod
	Currency       string
	AmountMinor    *int64
	StarsAmount    *int64
	AcceptedAssets *string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
