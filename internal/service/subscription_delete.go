package service

import (
	"context"
	"fmt"
	"strings"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}

	item, err := s.repo.Subscriptions.GetPublicByID(ctx, id)
	if err != nil {
		return err
	}

	if item.User.RemnaUUID != nil && strings.TrimSpace(*item.User.RemnaUUID) != "" {
		uuid := strings.TrimSpace(*item.User.RemnaUUID)
		if err := s.removeRemnaUserFromAllSquads(ctx, item); err != nil && !isIgnorableRemnaMissingError(err) {
			return err
		}
		if err := s.remna.DeleteUser(ctx, uuid); err != nil && !isIgnorableRemnaMissingError(err) {
			return fmt.Errorf("delete remnawave user before deleting subscription: %w", err)
		}
	}

	return s.repo.Subscriptions.Delete(ctx, id)
}
