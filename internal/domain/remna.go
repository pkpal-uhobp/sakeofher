package domain

type CreateRemnaUserRequest struct {
	Username          string `json:"username"`
	TrafficLimitBytes int64  `json:"traffic_limit_bytes"`
	ExpiresAtUnix     int64  `json:"expires_at_unix"`
}

type RemnaUser struct {
	UUID            string `json:"uuid"`
	Username        string `json:"username"`
	SubscriptionURL string `json:"subscription_url"`
	Status          string `json:"status"`
}

type RemnaTraffic struct {
	UsedBytes  int64 `json:"used_bytes"`
	LimitBytes int64 `json:"limit_bytes"`
}

type RemnaSyncAction string

const (
	RemnaSyncCreateUser   RemnaSyncAction = "create_user"
	RemnaSyncEnableUser   RemnaSyncAction = "enable_user"
	RemnaSyncDisableUser  RemnaSyncAction = "disable_user"
	RemnaSyncDeleteUser   RemnaSyncAction = "delete_user"
	RemnaSyncResetTraffic RemnaSyncAction = "reset_traffic"
	RemnaSyncUsage        RemnaSyncAction = "sync_usage"
)

type RemnaSyncLog struct {
	UserID          *int64          `json:"user_id,omitempty"`
	SubscriptionID  *int64          `json:"subscription_id,omitempty"`
	PaymentID       *int64          `json:"payment_id,omitempty"`
	Action          RemnaSyncAction `json:"action"`
	Success         bool            `json:"success"`
	ErrorText       *string         `json:"error_text,omitempty"`
	RequestPayload  []byte          `json:"request_payload,omitempty"`
	ResponsePayload []byte          `json:"response_payload,omitempty"`
}
