package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) ensureRemnaUserWithSquads(
	ctx context.Context,
	user *domain.User,
	trafficLimitBytes int64,
	expiresAt time.Time,
	activeInternalSquads []string,
) (*domain.RemnaUser, error) {
	username := remnaUsername(user)
	description := fmt.Sprintf("Telegram ID: %d", user.TelegramID)
	expiresAtUnix := expiresAt.Unix()
	squads := normalizeServiceSquads(activeInternalSquads)

	if user.RemnaUUID == nil || strings.TrimSpace(*user.RemnaUUID) == "" ||
		string(user.RemnaStatus) == "deleted" ||
		string(user.RemnaStatus) == "not_created" {
		return s.remna.CreateUser(ctx, domain.CreateRemnaUserRequest{
			Username:             username,
			TrafficLimitBytes:    trafficLimitBytes,
			ExpiresAtUnix:        expiresAtUnix,
			TrafficResetStrategy: "NO_RESET",
			Description:          description,
			TelegramID:           &user.TelegramID,
			ActiveInternalSquads: squads,
		})
	}

	// Do not call /actions/enable here.
	// For our "pause subscription" mode the Remnawave user usually remains ACTIVE,
	// but is removed from all internal squads. Calling EnableUser can fail or be
	// unnecessary. PATCH /api/users with Status=ACTIVE + squads is enough and
	// also refreshes the subscription config.
	return s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 strings.TrimSpace(*user.RemnaUUID),
		Username:             username,
		Status:               "ACTIVE",
		TrafficLimitBytes:    &trafficLimitBytes,
		ExpiresAtUnix:        &expiresAtUnix,
		TrafficResetStrategy: "NO_RESET",
		Description:          &description,
		TelegramID:           &user.TelegramID,
		ActiveInternalSquads: squads,
	})
}

func (s *subscriptionService) removeRemnaUserFromAllSquads(ctx context.Context, item *domain.PublicSubscription) error {
	if item == nil || item.User.RemnaUUID == nil || strings.TrimSpace(*item.User.RemnaUUID) == "" {
		return nil
	}

	description := fmt.Sprintf("Telegram ID: %d", item.User.TelegramID)
	expiresAtUnix := item.Subscription.ExpiresAt.Unix()
	trafficLimitBytes := item.Subscription.TrafficLimitBytes
	username := remnaUsername(&item.User)

	_, err := s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 strings.TrimSpace(*item.User.RemnaUUID),
		Username:             username,
		Status:               "ACTIVE",
		TrafficLimitBytes:    &trafficLimitBytes,
		ExpiresAtUnix:        &expiresAtUnix,
		TrafficResetStrategy: "NO_RESET",
		Description:          &description,
		TelegramID:           &item.User.TelegramID,

		// This must be sent as [] to Remnawave. Do not omit this field in DTO.
		ActiveInternalSquads: []string{},
	})
	if err != nil {
		return fmt.Errorf("remove remnawave user from squads: %w", err)
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

	return s.ensureRemnaUserWithSquads(
		ctx,
		&item.User,
		item.Subscription.TrafficLimitBytes,
		item.Subscription.ExpiresAt,
		squads,
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

	for _, item := range squads {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		if _, ok := seen[item]; ok {
			continue
		}

		seen[item] = struct{}{}
		out = append(out, item)
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
