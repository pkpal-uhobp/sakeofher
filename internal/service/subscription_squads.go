package service

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"sakeofher/internal/domain"
)

type remnaSquadUpdater interface {
	UpdateUserSquads(ctx context.Context, uuids []string, activeInternalSquads []string) error
}

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

	create := func() (*domain.RemnaUser, error) {
		remnaUser, err := s.remna.CreateUser(ctx, domain.CreateRemnaUserRequest{
			Username:              username,
			TrafficLimitBytes:     trafficLimitBytes,
			ExpiresAtUnix:         expiresAtUnix,
			TrafficResetStrategy:  "NO_RESET",
			Description:           description,
			TelegramID:            &user.TelegramID,
			ActiveInternalSquads:  squads,
		})
		if err != nil {
			return nil, err
		}
		return remnaUser, nil
	}

	if user.RemnaUUID == nil || strings.TrimSpace(*user.RemnaUUID) == "" || string(user.RemnaStatus) == "deleted" || string(user.RemnaStatus) == "not_created" {
		return create()
	}

	uuid := strings.TrimSpace(*user.RemnaUUID)
	if err := s.remna.EnableUser(ctx, uuid); err != nil {
		if isIgnorableRemnaMissingError(err) || isRemnaNotFoundError(err) {
			return create()
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
		TrafficResetStrategy: "NO_RESET",
		Description:          &description,
		TelegramID:           &user.TelegramID,
	})
	if err != nil {
		if isIgnorableRemnaMissingError(err) || isRemnaNotFoundError(err) {
			return create()
		}
		return nil, err
	}

	// Only assign squads when we actually have squads.
	// Do not call /api/users/bulk/update-squads with an empty list: some Remnawave
	// versions fail with A088 ("Bulk add inbounds to users error").
	if len(squads) > 0 {
		if err := s.syncRemnaUserSquads(ctx, remnaUser.UUID, squads); err != nil {
			if isIgnorableRemnaMissingError(err) || isRemnaNotFoundError(err) {
				return create()
			}
			return nil, err
		}
	}

	return remnaUser, nil
}

// Historical name kept for compatibility.
// Actual behavior now: disable user in Remnawave.
// Do not remove squads through /api/users/bulk/update-squads with an empty array:
// Remnawave may return A088 and the site/bot operation becomes HTTP 500.
func (s *subscriptionService) removeRemnaUserFromAllSquads(ctx context.Context, item *domain.PublicSubscription) error {
	if item == nil || item.User.RemnaUUID == nil || strings.TrimSpace(*item.User.RemnaUUID) == "" {
		return nil
	}

	uuid := strings.TrimSpace(*item.User.RemnaUUID)

	if err := s.remna.DisableUser(ctx, uuid); err != nil {
		if isIgnorableRemnaMissingError(err) || isRemnaNotFoundError(err) || isIgnorableRemnaAlreadyStateError(err) {
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

func (s *subscriptionService) syncRemnaUserSquads(ctx context.Context, remnaUUID string, squads []string) error {
	remnaUUID = strings.TrimSpace(remnaUUID)
	squads = normalizeServiceSquads(squads)
	if remnaUUID == "" || len(squads) == 0 {
		return nil
	}
	updater, ok := s.remna.(remnaSquadUpdater)
	if !ok {
		return nil
	}
	return updater.UpdateUserSquads(ctx, []string{remnaUUID}, squads)
}

func normalizeServiceSquads(squads []string) []string {
	out := make([]string, 0, len(squads))
	seen := make(map[string]struct{})
	for _, raw := range squads {
		for _, item := range strings.Split(raw, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			key := strings.ToLower(item)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	return out
}

func defaultRemnaSquadsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("BOT_REMNAWAVE_INTERNAL_SQUADS"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("REMNAWAVE_DEFAULT_INTERNAL_SQUADS"))
	}
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

func isIgnorableRemnaMissingError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status 404") ||
		strings.Contains(message, "not found") ||
		strings.Contains(message, "a025") ||
		strings.Contains(message, "a063")
}

func isRemnaNotFoundError(err error) bool {
	return isIgnorableRemnaMissingError(err)
}
