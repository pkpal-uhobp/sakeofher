package domain

import "time"

type SubscriptionLifecycleEventType string

const (
	SubscriptionLifecyclePaymentActivated   SubscriptionLifecycleEventType = "payment_activated"
	SubscriptionLifecycleManualCreated     SubscriptionLifecycleEventType = "manual_created"
	SubscriptionLifecycleRenewed           SubscriptionLifecycleEventType = "renewed"
	SubscriptionLifecycleExtended          SubscriptionLifecycleEventType = "extended"
	SubscriptionLifecycleDisabled          SubscriptionLifecycleEventType = "disabled"
	SubscriptionLifecycleEnabled           SubscriptionLifecycleEventType = "enabled"
	SubscriptionLifecycleCancelled         SubscriptionLifecycleEventType = "cancelled"
	SubscriptionLifecycleExpired           SubscriptionLifecycleEventType = "expired"
	SubscriptionLifecycleTrafficExhausted  SubscriptionLifecycleEventType = "traffic_exhausted"
	SubscriptionLifecycleTrafficReset      SubscriptionLifecycleEventType = "traffic_reset"
	SubscriptionLifecycleRemnaReconciled   SubscriptionLifecycleEventType = "remna_reconciled"
	SubscriptionLifecycleRemnaSyncFailed   SubscriptionLifecycleEventType = "remna_sync_failed"
)

type SubscriptionLifecycleEvent struct {
	ID               int64                          `json:"id"`
	SubscriptionID   *int64                         `json:"subscription_id,omitempty"`
	UserID           *int64                         `json:"user_id,omitempty"`
	PaymentID        *int64                         `json:"payment_id,omitempty"`
	EventType        SubscriptionLifecycleEventType `json:"event_type"`
	FromStatus       *SubscriptionStatus            `json:"from_status,omitempty"`
	ToStatus         *SubscriptionStatus            `json:"to_status,omitempty"`
	FromPeriodStatus *PeriodStatus                  `json:"from_period_status,omitempty"`
	ToPeriodStatus   *PeriodStatus                  `json:"to_period_status,omitempty"`
	Reason           string                         `json:"reason,omitempty"`
	Success          bool                           `json:"success"`
	ErrorText         *string                        `json:"error_text,omitempty"`
	Details          []byte                         `json:"details,omitempty"`
	CreatedAt        time.Time                      `json:"created_at"`
}

type DesiredRemnaState struct {
	Enabled              bool     `json:"enabled"`
	Status               string   `json:"status"`
	ActiveInternalSquads []string `json:"active_internal_squads"`
	Reason               string   `json:"reason"`
}
