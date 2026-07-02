package domain

import "time"

type Subscription struct {
	ID                 int64
	UserID             int64
	TariffID           int64
	Status             SubscriptionStatus
	StartedAt          time.Time
	ExpiresAt          time.Time
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	TrafficLimitBytes  int64
	TrafficUsedBytes   int64
	Notified80         bool
	Notified95         bool
	Notified3Days      bool
	Notified1Day       bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
