package domain

import "time"

type CreateRemnaUserRequest struct {
	Username          string   `json:"username"`
	TrafficLimitBytes int64    `json:"traffic_limit_bytes"`
	ExpiresAtUnix     int64    `json:"expires_at_unix"`
	TrafficResetStrategy string `json:"traffic_reset_strategy,omitempty"`
	Description       string   `json:"description,omitempty"`
	TelegramID        *int64   `json:"telegram_id,omitempty"`
	Email             *string  `json:"email,omitempty"`
	Tag               *string  `json:"tag,omitempty"`
	ActiveInternalSquads []string `json:"active_internal_squads,omitempty"`
}

type UpdateRemnaUserRequest struct {
	UUID              string   `json:"uuid"`
	Username          string   `json:"username,omitempty"`
	Status            string   `json:"status,omitempty"`
	TrafficLimitBytes *int64   `json:"traffic_limit_bytes,omitempty"`
	ExpiresAtUnix     *int64   `json:"expires_at_unix,omitempty"`
	TrafficResetStrategy string `json:"traffic_reset_strategy,omitempty"`
	Description       *string  `json:"description,omitempty"`
	TelegramID        *int64   `json:"telegram_id,omitempty"`
	Email             *string  `json:"email,omitempty"`
	Tag               *string  `json:"tag,omitempty"`
	ActiveInternalSquads []string `json:"active_internal_squads,omitempty"`
}

type RemnaUser struct {
	UUID                 string     `json:"uuid"`
	ShortUUID            string     `json:"short_uuid,omitempty"`
	Username             string     `json:"username"`
	SubscriptionURL      string     `json:"subscription_url"`
	Status               string     `json:"status"`
	UsedTrafficBytes     int64      `json:"used_traffic_bytes"`
	LifetimeUsedTrafficBytes int64  `json:"lifetime_used_traffic_bytes"`
	TrafficLimitBytes    int64      `json:"traffic_limit_bytes"`
	TrafficLimitStrategy string     `json:"traffic_limit_strategy,omitempty"`
	ExpireAt            *time.Time `json:"expire_at,omitempty"`
	LastTrafficResetAt   *time.Time `json:"last_traffic_reset_at,omitempty"`
}

type RemnaTraffic struct {
	UsedBytes  int64 `json:"used_bytes"`
	LimitBytes int64 `json:"limit_bytes"`
}

type RemnaSyncAction string

const (
	RemnaSyncCreateUser    RemnaSyncAction = "create_user"
	RemnaSyncUpdateUser    RemnaSyncAction = "update_user"
	RemnaSyncEnableUser    RemnaSyncAction = "enable_user"
	RemnaSyncDisableUser   RemnaSyncAction = "disable_user"
	RemnaSyncDeleteUser    RemnaSyncAction = "delete_user"
	RemnaSyncResetTraffic  RemnaSyncAction = "reset_traffic"
	RemnaSyncUsage         RemnaSyncAction = "sync_usage"
)

type RemnaSyncLog struct {
	UserID          *int64          `json:"user_id,omitempty"`
	SubscriptionID *int64          `json:"subscription_id,omitempty"`
	PaymentID      *int64          `json:"payment_id,omitempty"`
	Action          RemnaSyncAction `json:"action"`
	Success         bool            `json:"success"`
	ErrorText       *string         `json:"error_text,omitempty"`
	RequestPayload  []byte          `json:"request_payload,omitempty"`
	ResponsePayload []byte          `json:"response_payload,omitempty"`
}
