package service

import (
	"context"

	"sakeofher/internal/domain"
)

func (s *subscriptionService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return domain.ErrInvalidInput
	}

	return s.repo.Subscriptions.Delete(ctx, id)
}
