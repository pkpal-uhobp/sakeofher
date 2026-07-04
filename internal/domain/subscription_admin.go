package domain

import "time"

type SubscriptionListInput struct {
	UserID     int64              `json:"user_id"`
	TelegramID int64              `json:"telegram_id"`
	Status     SubscriptionStatus `json:"status"`
	Limit      int                `json:"limit"`
	Offset     int                `json:"offset"`
}

type SubscriptionListResponse struct {
	Items  []PublicSubscription `json:"items"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

type CreateManualSubscriptionInput struct {
	UserID               int64    `json:"user_id"`
	TariffID             int64    `json:"tariff_id"`
	TrafficLimitGB       int64    `json:"traffic_limit_gb"`
	ActiveInternalSquads []string `json:"active_internal_squads,omitempty"`
}

type ExtendSubscriptionInput struct {
	TariffID             *int64   `json:"tariff_id,omitempty"`
	Days                 *int     `json:"days,omitempty"`
	ActiveInternalSquads []string `json:"active_internal_squads,omitempty"`
}

type UpdateTrafficLimitInput struct {
	TrafficLimitGB int64 `json:"traffic_limit_gb"`
}

type UpdateSubscriptionInput struct {
	Status             *SubscriptionStatus `json:"status,omitempty"`
	PeriodStatus       *PeriodStatus       `json:"period_status,omitempty"`
	ExpiresAt          *time.Time          `json:"expires_at,omitempty"`
	CurrentPeriodStart *time.Time          `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time          `json:"current_period_end,omitempty"`
	TrafficLimitGB      *int64              `json:"traffic_limit_gb,omitempty"`
	TrafficUsedGB       *int64              `json:"traffic_used_gb,omitempty"`
}
