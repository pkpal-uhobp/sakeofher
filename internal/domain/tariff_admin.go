package domain

type TariffListInput struct {
	OnlyActive bool `json:"only_active"`
}

type CreateTariffInput struct {
	Code           string                      `json:"code"`
	Title          string                      `json:"title"`
	Description    *string                     `json:"description,omitempty"`
	DurationDays   int                         `json:"duration_days"`
	PeriodDays     int                         `json:"period_days"`
	TrafficLimitGB int64                       `json:"traffic_limit_gb"`
	PriceRub       int64                       `json:"price_rub"`
	IsActive       *bool                       `json:"is_active,omitempty"`
	SortOrder      int                         `json:"sort_order"`
	PaymentSettings *TariffPaymentSettingsInput `json:"payment_settings,omitempty"`
}

type UpdateTariffInput struct {
	Code           *string                     `json:"code,omitempty"`
	Title          *string                     `json:"title,omitempty"`
	Description    *string                     `json:"description,omitempty"`
	DurationDays   *int                        `json:"duration_days,omitempty"`
	PeriodDays     *int                        `json:"period_days,omitempty"`
	TrafficLimitGB *int64                      `json:"traffic_limit_gb,omitempty"`
	PriceRub       *int64                      `json:"price_rub,omitempty"`
	IsActive       *bool                       `json:"is_active,omitempty"`
	SortOrder      *int                        `json:"sort_order,omitempty"`
	PaymentSettings *TariffPaymentSettingsInput `json:"payment_settings,omitempty"`
}

type TariffPaymentSettingsInput struct {
	TelegramStars TariffTelegramStarsSettings   `json:"telegram_stars"`
	CryptoBotCrypto TariffCryptoBotCryptoSettings `json:"cryptobot_crypto"`
	TributeRub    TariffTributeRubSettings      `json:"tribute_rub"`
}

type TariffTelegramStarsSettings struct {
	Enabled     bool  `json:"enabled"`
	StarsAmount int64 `json:"stars_amount"`
}

type TariffCryptoBotCryptoSettings struct {
	Enabled        bool     `json:"enabled"`
	PriceRub       int64    `json:"price_rub"`
	AcceptedAssets []string `json:"accepted_assets"`
}

type TariffTributeRubSettings struct {
	Enabled  bool  `json:"enabled"`
	PriceRub int64 `json:"price_rub"`
}
