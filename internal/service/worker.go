package service

import "context"

type WorkerService struct {
	subscriptions *SubscriptionService
	payments      *PaymentService
}

func NewWorkerService(subscriptions *SubscriptionService, payments *PaymentService) *WorkerService {
	return &WorkerService{subscriptions: subscriptions, payments: payments}
}

func (s *WorkerService) ExpireSubscriptions(ctx context.Context) error {
	return s.subscriptions.DisableExpiredSubscriptions(ctx, 100)
}
