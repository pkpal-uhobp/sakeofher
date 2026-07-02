package domain

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

type RemnaStatus string

const (
	RemnaStatusNotCreated RemnaStatus = "not_created"
	RemnaStatusActive     RemnaStatus = "active"
	RemnaStatusDisabled   RemnaStatus = "disabled"
	RemnaStatusDeleted    RemnaStatus = "deleted"
)

type PaymentProvider string

const (
	PaymentProviderTelegramStars PaymentProvider = "telegram_stars"
	PaymentProviderTribute       PaymentProvider = "tribute"
	PaymentProviderCryptoBot     PaymentProvider = "crypto_bot"
)

type PaymentMethod string

const (
	PaymentMethodStars  PaymentMethod = "stars"
	PaymentMethodRub    PaymentMethod = "rub"
	PaymentMethodCrypto PaymentMethod = "crypto"
)

type PaymentStatus string

const (
	PaymentStatusCreated          PaymentStatus = "created"
	PaymentStatusWaitingPayment   PaymentStatus = "waiting_payment"
	PaymentStatusPaid             PaymentStatus = "paid"
	PaymentStatusActivationFailed PaymentStatus = "activation_failed"
	PaymentStatusActivated        PaymentStatus = "activated"
	PaymentStatusFailed           PaymentStatus = "failed"
	PaymentStatusCancelled        PaymentStatus = "cancelled"
	PaymentStatusExpired          PaymentStatus = "expired"
	PaymentStatusRefunded         PaymentStatus = "refunded"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusExpired   SubscriptionStatus = "expired"
	SubscriptionStatusCancelled SubscriptionStatus = "cancelled"
)

type PeriodStatus string

const (
	PeriodStatusActive           PeriodStatus = "active"
	PeriodStatusTrafficExhausted PeriodStatus = "traffic_exhausted"
	PeriodStatusFinished         PeriodStatus = "finished"
)

type BroadcastStatus string

const (
	BroadcastStatusDraft     BroadcastStatus = "draft"
	BroadcastStatusQueued    BroadcastStatus = "queued"
	BroadcastStatusSending   BroadcastStatus = "sending"
	BroadcastStatusFinished  BroadcastStatus = "finished"
	BroadcastStatusFailed    BroadcastStatus = "failed"
	BroadcastStatusCancelled BroadcastStatus = "cancelled"
)
