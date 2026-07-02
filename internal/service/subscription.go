package service

import (
	"context"
	"time"

	"sakeofher/internal/gateway"
	"sakeofher/internal/repository"
)

type SubscriptionService struct {
	repo          *repository.Repositories
	remna         gateway.RemnawaveGateway
	notifications *NotificationService
}

func NewSubscriptionService(repo *repository.Repositories, remna gateway.RemnawaveGateway, notifications *NotificationService) *SubscriptionService {
	return &SubscriptionService{repo: repo, remna: remna, notifications: notifications}
}

func (s *SubscriptionService) DisableExpiredSubscriptions(ctx context.Context, limit int) error {
	expired, err := s.repo.Subscriptions.FindExpiredActive(ctx, time.Now(), limit)
	if err != nil {
		return err
	}
	for _, sub := range expired {
		_ = sub // TODO: load user, call Remnawave DisableUser, update statuses in transaction.
	}
	return nil
}
