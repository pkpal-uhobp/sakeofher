package domain

type CreateRemnaUserRequest struct {
	Username          string
	TrafficLimitBytes int64
	ExpiresAtUnix     int64
}

type RemnaUser struct {
	UUID            string
	Username        string
	SubscriptionURL string
	Status          string
}

type RemnaTraffic struct {
	UsedBytes  int64
	LimitBytes int64
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
