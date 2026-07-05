package service

import (
	"context"
	"fmt"
	"time"

	"sakeofher/internal/domain"
)

// AdvanceTrafficExhaustedPeriods moves a subscription to the next paid traffic period
// when the current period quota is exhausted before current_period_end.
//
// This is important for multi-month tariffs:
// duration_days=90, period_days=30, traffic_limit=100GB means 3 paid traffic periods.
// If the user spends 100GB in the first 10 days, we reset usage and immediately start
// the next paid period. The final subscription expires_at is not extended.
func (s *subscriptionService) AdvanceTrafficExhaustedPeriods(ctx context.Context, limit int) error {
	now := time.Now()
	items, err := s.repo.Subscriptions.FindTrafficExhaustedReadyForAdvance(ctx, now, limit)
	if err != nil {
		return err
	}

	for _, item := range items {
		periodDays := item.Tariff.PeriodDays
		if periodDays <= 0 {
			periodDays = 30
		}

		nextStart := now
		nextEnd := nextStart.AddDate(0, 0, periodDays)
		if nextEnd.After(item.Subscription.ExpiresAt) {
			nextEnd = item.Subscription.ExpiresAt
		}

		// No paid periods are left. The normal expiration flow will disable the user.
		if !nextEnd.After(nextStart) || !item.Subscription.ExpiresAt.After(now) {
			continue
		}

		if item.User.RemnaUUID != nil && *item.User.RemnaUUID != "" {
			if err := s.remna.ResetTraffic(ctx, *item.User.RemnaUUID); err != nil {
				// If the user was manually deleted in Remnawave, restoring below will create
				// a fresh Remnawave user with zero traffic. Other errors must stop the job.
				if !isRemnaNotFoundError(err) {
					return fmt.Errorf("reset remnawave traffic before exhausted-period advance: %w", err)
				}
			}
		}

		publicItem := &domain.PublicSubscription{
			Subscription: item.Subscription,
			User:         item.User,
			Tariff:       item.Tariff,
		}
		remnaUser, err := s.restoreRemnaUserSquads(ctx, publicItem, nil)
		if err != nil {
			return fmt.Errorf("restore remnawave user after exhausted-period advance: %w", err)
		}
		if remnaUser != nil {
			if err := s.repo.Users.SetRemnaData(ctx, item.User.ID, domain.RemnaUserData{
				UUID:            remnaUser.UUID,
				Username:        remnaUser.Username,
				SubscriptionURL: remnaUser.SubscriptionURL,
				Status:          domain.RemnaStatus("active"),
			}); err != nil {
				return err
			}
		}

		if err := s.repo.Subscriptions.AdvanceTrafficPeriodAfterExhaustion(ctx, item.Subscription.ID, nextStart, nextEnd); err != nil {
			return err
		}

		// Do not notify the user here. If the subscription still has paid traffic
		// periods left, rollover must be invisible: traffic is reset and the next
		// paid period starts without a warning or renewal prompt.
	}

	return nil
}
