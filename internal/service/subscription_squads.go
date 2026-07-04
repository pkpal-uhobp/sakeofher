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

	uuid := strings.TrimSpace(*user.RemnaUUID)

	if err := s.remna.EnableUser(ctx, uuid); err != nil && !isIgnorableRemnaAlreadyStateError(err) {
		return nil, fmt.Errorf("enable remnawave user: %w", err)
	}

	return s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 uuid,
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

// Historical name kept for compatibility. Actual behavior:
// remove user from squads and disable him in Remnawave.
func (s *subscriptionService) removeRemnaUserFromAllSquads(ctx context.Context, item *domain.PublicSubscription) error {
	if item == nil || item.User.RemnaUUID == nil || strings.TrimSpace(*item.User.RemnaUUID) == "" {
		return nil
	}

	uuid := strings.TrimSpace(*item.User.RemnaUUID)
	description := fmt.Sprintf("Telegram ID: %d; subscription expired or disabled", item.User.TelegramID)
	expiresAtUnix := item.Subscription.ExpiresAt.Unix()
	trafficLimitBytes := item.Subscription.TrafficLimitBytes
	username := remnaUsername(&item.User)

	if _, err := s.remna.UpdateUser(ctx, domain.UpdateRemnaUserRequest{
		UUID:                 uuid,
		Username:             username,
		Status:               "ACTIVE",
		TrafficLimitBytes:    &trafficLimitBytes,
		ExpiresAtUnix:        &expiresAtUnix,
		TrafficResetStrategy: "NO_RESET",
		Description:          &description,
		TelegramID:           &item.User.TelegramID,
		ActiveInternalSquads: []string{},
	}); err != nil {
		return fmt.Errorf("remove remnawave user from squads before disable: %w", err)
	}

	if err := s.remna.DisableUser(ctx, uuid); err != nil && !isIgnorableRemnaAlreadyStateError(err) {
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
