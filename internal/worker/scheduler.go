package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"sakeofher/internal/config"
	"sakeofher/internal/service"
)

type Scheduler struct {
	services *service.Services
	cfg      config.WorkerConfig
	log      *zap.Logger
}

func NewScheduler(services *service.Services, cfg config.WorkerConfig, log *zap.Logger) *Scheduler {
	return &Scheduler{services: services, cfg: cfg, log: log}
}

func (s *Scheduler) Run(ctx context.Context) error {
	expireTicker := time.NewTicker(s.cfg.ExpireInterval)
	deleteTicker := time.NewTicker(s.cfg.DeleteDisabledInterval)
	retryTicker := time.NewTicker(s.cfg.RetryActivationInterval)
	defer expireTicker.Stop()
	defer deleteTicker.Stop()
	defer retryTicker.Stop()

	s.log.Info("worker scheduler started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-expireTicker.C:
			if err := s.services.Workers.ExpireSubscriptions(ctx); err != nil {
				s.log.Error("expire subscriptions failed", zap.Error(err))
			}
		case <-deleteTicker.C:
			if err := s.services.Workers.DeleteOldDisabledUsers(ctx); err != nil {
				s.log.Error("delete old disabled users failed", zap.Error(err))
			}
		case <-retryTicker.C:
			if err := s.services.Workers.RetryFailedActivations(ctx); err != nil {
				s.log.Error("retry failed activations failed", zap.Error(err))
			}
		}
	}
}
