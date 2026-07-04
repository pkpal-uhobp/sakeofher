package domain

type SubscriptionWithUserAndTariff struct {
	Subscription Subscription `json:"subscription"`
	User         User         `json:"user"`
	Tariff       Tariff       `json:"tariff"`
}
