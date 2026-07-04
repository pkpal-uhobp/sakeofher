package service

import "context"

type workerService struct {
	subscriptions SubscriptionService
	payments      PaymentService
}

func NewWorkerService(subscriptions SubscriptionService, payments PaymentService) WorkerService {
	return &workerService{subscriptions: subscriptions, payments: payments}
}

func (s *workerService) ExpireSubscriptions(ctx context.Context) error {
	return s.subscriptions.DisableExpiredSubscriptions(ctx, 100)
}

func (s *workerService) DeleteOldDisabledUsers(ctx context.Context) error {
	return s.subscriptions.DeleteOldDisabledUsers(ctx, 100)
}

func (s *workerService) RetryFailedActivations(ctx context.Context) error {
	return s.payments.RetryFailedActivations(ctx, 50)
}

func (s *workerService) SyncUsage(ctx context.Context) error {
	return s.subscriptions.SyncRemnaUsage(ctx, 100)
}

func (s *workerService) ResetTrafficPeriods(ctx context.Context) error {
	return s.subscriptions.ResetTrafficPeriods(ctx, 100)
}
