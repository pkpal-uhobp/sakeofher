package domain

type TariffListInput struct {
	OnlyActive bool `json:"only_active"`
}

type CreateTariffInput struct {
	Code              string  `json:"code"`
	Title             string  `json:"title"`
	Description       *string `json:"description,omitempty"`
	DurationDays      int     `json:"duration_days"`
	PeriodDays        int     `json:"period_days"`
	TrafficLimitGB    int64   `json:"traffic_limit_gb"`
	IsActive          *bool   `json:"is_active,omitempty"`
	SortOrder         int     `json:"sort_order"`
}

type UpdateTariffInput struct {
	Code              *string `json:"code,omitempty"`
	Title             *string `json:"title,omitempty"`
	Description       *string `json:"description,omitempty"`
	DurationDays      *int    `json:"duration_days,omitempty"`
	PeriodDays        *int    `json:"period_days,omitempty"`
	TrafficLimitGB    *int64  `json:"traffic_limit_gb,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
	SortOrder         *int    `json:"sort_order,omitempty"`
}
