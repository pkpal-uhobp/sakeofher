package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) ReconcileRemnaState(ctx context.Context, limit int) error {
	now := time.Now()

	items, err := s.repo.Subscriptions.FindRemnaReconcileCandidates(ctx, now, limit)
	if err != nil {
		return err
	}

	for _, item := range items {
		desired, err := s.desiredRemnaState(ctx, &item, now)
		if err != nil {
			return err
		}

		if err := s.applyDesiredRemnaState(ctx, &item, desired, now); err != nil {
			errorText := err.Error()
			_ = s.repo.Users.MarkRemnaSquadsSyncFailed(ctx, item.User.ID, desired.ActiveInternalSquads, errorText)
			_ = s.writeLifecycleEvent(ctx, &item, domain.SubscriptionLifecycleRemnaSyncFailed, desired.Reason, false, &errorText, desired)
			return err
		}

		_ = s.writeLifecycleEvent(ctx, &item, domain.SubscriptionLifecycleRemnaReconciled, desired.Reason, true, nil, desired)
	}

	return nil
}

func (s *subscriptionService) desiredRemnaState(
	ctx context.Context,
	item *domain.PublicSubscription,
	now time.Time,
) (domain.DesiredRemnaState, error) {
	if item == nil {
		return domain.DesiredRemnaState{}, domain.ErrNotFound
	}

	status := item.Subscription.Status
	periodStatus := item.Subscription.PeriodStatus

	if status == domain.SubscriptionStatusExpired ||
		status == domain.SubscriptionStatusCancelled ||
		periodStatus == domain.PeriodStatusFinished ||
		periodStatus == domain.PeriodStatusTrafficExhausted ||
		item.Subscription.ExpiresAt.Before(now) ||
		item.Subscription.ExpiresAt.Equal(now) {
		return domain.DesiredRemnaState{
			Enabled:              false,
			Status:               "DISABLED",
			ActiveInternalSquads: []string{},
			Reason:               "subscription_not_active",
		}, nil
	}

	squads, err := s.repo.Users.GetRemnaActiveSquads(ctx, item.User.ID)
	if err != nil {
		return domain.DesiredRemnaState{}, err
	}

	if len(squads) == 0 {
		squads = defaultRemnaSquadsFromEnv()
	}

	return domain.DesiredRemnaState{
		Enabled:              true,
		Status:               "ACTIVE",
		ActiveInternalSquads: normalizeServiceSquads(squads),
		Reason:               "subscription_active",
	}, nil
}

func (s *subscriptionService) applyDesiredRemnaState(
	ctx context.Context,
	item *domain.PublicSubscription,
	desired domain.DesiredRemnaState,
	now time.Time,
) error {
	if item == nil {
		return domain.ErrNotFound
	}

	if desired.Enabled {
		remnaUser, err := s.restoreRemnaUserSquads(ctx, item, desired.ActiveInternalSquads)
		if err != nil {
			return err
		}

		if err := s.repo.Users.SetRemnaData(ctx, item.User.ID, domain.RemnaUserData{
			UUID:            remnaUser.UUID,
			Username:        remnaUser.Username,
			SubscriptionURL: remnaUser.SubscriptionURL,
			Status:          domain.RemnaStatusActive,
		}); err != nil {
			return err
		}

		if err := s.repo.Users.MarkRemnaSquadsSynced(ctx, item.User.ID, desired.ActiveInternalSquads, now); err != nil {
			return err
		}

		return nil
	}

	if err := s.removeRemnaUserFromAllSquads(ctx, item); err != nil {
		return err
	}

	deleteAfter := now.AddDate(0, 0, 7)

	if item.Subscription.Status == domain.SubscriptionStatusActive &&
		(item.Subscription.ExpiresAt.Before(now) || item.Subscription.ExpiresAt.Equal(now)) {
		if err := s.repo.Subscriptions.MarkExpired(ctx, item.Subscription.ID); err != nil {
			return err
		}
	}

	if err := s.repo.Users.MarkRemnaDisabled(ctx, item.User.ID, now, deleteAfter); err != nil {
		return err
	}

	if err := s.repo.Users.MarkRemnaSquadsSynced(ctx, item.User.ID, []string{}, now); err != nil {
		return err
	}

	return nil
}

func (s *subscriptionService) writeLifecycleEvent(
	ctx context.Context,
	item *domain.PublicSubscription,
	eventType domain.SubscriptionLifecycleEventType,
	reason string,
	success bool,
	errorText *string,
	details any,
) error {
	if item == nil {
		return nil
	}

	subscriptionID := item.Subscription.ID
	userID := item.User.ID
	fromStatus := item.Subscription.Status
	toStatus := item.Subscription.Status
	fromPeriodStatus := item.Subscription.PeriodStatus
	toPeriodStatus := item.Subscription.PeriodStatus

	rawDetails, _ := json.Marshal(details)

	return s.repo.Subscriptions.CreateLifecycleEvent(ctx, domain.SubscriptionLifecycleEvent{
		SubscriptionID:   &subscriptionID,
		UserID:           &userID,
		EventType:        eventType,
		FromStatus:       &fromStatus,
		ToStatus:         &toStatus,
		FromPeriodStatus: &fromPeriodStatus,
		ToPeriodStatus:   &toPeriodStatus,
		Reason:           reason,
		Success:          success,
		ErrorText:         errorText,
		Details:          rawDetails,
	})
}

func remnaSyncError(action string, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", action, err)
}
