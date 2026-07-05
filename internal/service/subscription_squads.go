package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

const (
	remnaTrafficStrategyNoReset = "NO_RESET"
	remnaTrafficStrategyDay     = "DAY"
	remnaTrafficStrategyWeek    = "WEEK"
	remnaTrafficStrategyMonth   = "MONTH"
)

func (s *subscriptionService) ensureRemnaUserWithSquads(
	ctx context.Context,
	user *domain.User,
	trafficLimitBytes int64,
	expiresAt time.Time,
	activeInternalSquads []string,
) (*domain.RemnaUser, error) {
	return s.ensureRemnaUserWithSquadsAndStrategy(
		ctx,
		user,
		trafficLimitBytes,
		expiresAt,
		activeInternalSquads,
		remnaTrafficStrategyNoReset,
	)
}

func (s *subscriptionService) ensureRemnaUserWithSquadsAndStrategy(
	ctx context.Context,
	user *domain.User,
	trafficLimitBytes int64,
	expiresAt time.Time,
	activeInternalSquads []string,
	trafficResetStrategy string,
) (*domain.RemnaUser, error) {
	username := remnaUsername(user)
	description := fmt.Sprintf("Telegram ID: %d", user.TelegramID)
	expiresAtUnix := expiresAt.Unix()
	squads := normalizeServiceSquads(activeInternalSquads)
	strategy := normalizeRemnaTrafficResetStrategy(trafficResetStrategy)

	createReq := domain.CreateRemnaUserRequest{
		Username:             username,
		TrafficLimitBytes:    trafficLimitBytes,
		ExpiresAtUnix:        expiresAtUnix,
		TrafficResetStrategy: strategy,
		Description:          description,
		TelegramID:           &user.TelegramID,
		ActiveInternalSquads: squads,
	}

	if user.RemnaUUID == nil || strings.TrimSpace(*user.RemnaUUID) == "" || string(user.RemnaStatus) == "deleted" || string(user.RemnaStatus) == "not_created" {
		return s.remna.CreateUser(ctx, createReq)
	}

	uuid := strings.TrimSpace(*user.RemnaUUID)
	if err := s.remna.EnableUser(ctx, uuid); err != nil {
		// If Remnawave was wiped or the user was deleted manually from the panel,
		// the site DB still keeps the old UUID. Do not fail the worker forever:
		// create a fresh Remnawave user and let the caller save the new UUID/URL.
		if isRemnaNotFoundError(err) {
			return s.remna.CreateUser(ctx, createReq)
		}
		if !isIgnorableRemnaAlreadyStateError(err) {
			return nil, fmt.Errorf("enable remnawave user: %w", err)
		}
	}

	remnaUser, err := s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 uuid,
		Username:             username,
		Status:               "ACTIVE",
		TrafficLimitBytes:    &trafficLimitBytes,
		ExpiresAtUnix:        &expiresAtUnix,
		TrafficResetStrategy: strategy,
		Description:          &description,
		TelegramID:           &user.TelegramID,
		ActiveInternalSquads: squads,
	})
	if err != nil {
		if isRemnaNotFoundError(err) {
			return s.remna.CreateUser(ctx, createReq)
		}
		return nil, err
	}
	return remnaUser, nil
}

// Historical name kept for compatibility.
// Actual behavior: remove user from squads and disable him in Remnawave.
func (s *subscriptionService) removeRemnaUserFromAllSquads(ctx context.Context, item *domain.PublicSubscription) error {
	if item == nil || item.User.RemnaUUID == nil || strings.TrimSpace(*item.User.RemnaUUID) == "" {
		return nil
	}

	uuid := strings.TrimSpace(*item.User.RemnaUUID)
	description := fmt.Sprintf("Telegram ID: %d; subscription expired or disabled", item.User.TelegramID)
	expiresAtUnix := item.Subscription.ExpiresAt.Unix()
	trafficLimitBytes := item.Subscription.TrafficLimitBytes
	username := remnaUsername(&item.User)
	strategy := remnaTrafficResetStrategyForTariff(item.Tariff)

	if _, err := s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 uuid,
		Username:             username,
		Status:               "ACTIVE",
		TrafficLimitBytes:    &trafficLimitBytes,
		ExpiresAtUnix:        &expiresAtUnix,
		TrafficResetStrategy: strategy,
		Description:          &description,
		TelegramID:           &item.User.TelegramID,
		ActiveInternalSquads: []string{},
	}); err != nil {
		if isRemnaNotFoundError(err) {
			return nil
		}
		return fmt.Errorf("remove remnawave user from squads before disable: %w", err)
	}
	if err := s.remna.DisableUser(ctx, uuid); err != nil {
		if isRemnaNotFoundError(err) || isIgnorableRemnaAlreadyStateError(err) {
			return nil
		}
		return fmt.Errorf("disable remnawave user: %w", err)
	}
	return nil
}

func (s *subscriptionService) restoreRemnaUserSquads(
	ctx context.Context,
	item *domain.PublicSubscription,
	overrideSquads []string,
) (*domain.RemnaUser, error) {
	if item == nil {
		return nil, domain.ErrNotFound
	}

	squads := normalizeServiceSquads(overrideSquads)
	if len(squads) == 0 {
		saved, err := s.repo.Users.GetRemnaActiveSquads(ctx, item.User.ID)
		if err != nil {
			return nil, err
		}
		squads = saved
	}
	if len(squads) == 0 {
		squads = defaultRemnaSquadsFromEnv()
	}

	return s.ensureRemnaUserWithSquadsAndStrategy(
		ctx,
		&item.User,
		item.Subscription.TrafficLimitBytes,
		item.Subscription.ExpiresAt,
		squads,
		remnaTrafficResetStrategyForTariff(item.Tariff),
	)
}

func (s *subscriptionService) savePreferredRemnaSquads(ctx context.Context, userID int64, squads []string) error {
	squads = normalizeServiceSquads(squads)
	if len(squads) == 0 {
		return nil
	}
	return s.repo.Users.SetRemnaActiveSquads(ctx, userID, squads)
}

func normalizeServiceSquads(squads []string) []string {
	out := make([]string, 0, len(squads))
	seen := make(map[string]struct{})
	for _, raw := range squads {
		for _, part := range strings.Split(raw, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			key := strings.ToLower(part)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, part)
		}
	}
	return out
}

func defaultRemnaSquadsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("REMNAWAVE_DEFAULT_INTERNAL_SQUADS"))
	if raw == "" {
		return nil
	}
	return normalizeServiceSquads(strings.Split(raw, ","))
}

func remnaTrafficResetStrategyForTariff(_ domain.Tariff) string {
	// The site DB is the source of truth for paid traffic periods.
	// For multi-month tariffs we may advance to the next paid period immediately
	// when traffic is exhausted, so Remnawave must not perform an extra calendar
	// reset on its own. The worker calls /actions/reset-traffic exactly when the
	// site period rolls over.
	return remnaTrafficStrategyNoReset
}

func remnaTrafficResetStrategy(periodDays int) string {
	switch {
	case periodDays == 1:
		return remnaTrafficStrategyDay
	case periodDays >= 6 && periodDays <= 8:
		return remnaTrafficStrategyWeek
	case periodDays >= 28 && periodDays <= 31:
		return remnaTrafficStrategyMonth
	default:
		return remnaTrafficStrategyNoReset
	}
}

func normalizeRemnaTrafficResetStrategy(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case remnaTrafficStrategyDay:
		return remnaTrafficStrategyDay
	case remnaTrafficStrategyWeek:
		return remnaTrafficStrategyWeek
	case remnaTrafficStrategyMonth:
		return remnaTrafficStrategyMonth
	default:
		return remnaTrafficStrategyNoReset
	}
}

func isRemnaNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status 404") ||
		strings.Contains(message, "user not found") ||
		strings.Contains(message, "not found") ||
		strings.Contains(message, "a025") ||
		strings.Contains(message, "a063")
}

func isIgnorableRemnaAlreadyStateError(err error) bool {
	if err == nil {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "already") ||
		strings.Contains(message, "same status") ||
		strings.Contains(message, "not modified") ||
		strings.Contains(message, "no changes")
}
