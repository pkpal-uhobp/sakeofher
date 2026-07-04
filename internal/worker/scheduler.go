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
	syncUsageTicker := time.NewTicker(s.cfg.SyncUsageInterval)
	resetTrafficTicker := time.NewTicker(s.cfg.ResetTrafficInterval)
	notifyTicker := time.NewTicker(s.cfg.NotifyInterval)

	defer expireTicker.Stop()
	defer deleteTicker.Stop()
	defer retryTicker.Stop()
	defer syncUsageTicker.Stop()
	defer resetTrafficTicker.Stop()
	defer notifyTicker.Stop()

	s.log.Info(
		"worker scheduler started",
		zap.Duration("expire_interval", s.cfg.ExpireInterval),
		zap.Duration("delete_disabled_interval", s.cfg.DeleteDisabledInterval),
		zap.Duration("retry_activation_interval", s.cfg.RetryActivationInterval),
		zap.Duration("sync_usage_interval", s.cfg.SyncUsageInterval),
		zap.Duration("reset_traffic_interval", s.cfg.ResetTrafficInterval),
		zap.Duration("notify_interval", s.cfg.NotifyInterval),
	)

	// Run immediately, not after the first hour.
	s.runJob(ctx, "expire subscriptions", s.services.Workers.ExpireSubscriptions)
	s.runJob(ctx, "sync remnawave usage", s.services.Workers.SyncUsage)
	s.runJob(ctx, "reset traffic periods", s.services.Workers.ResetTrafficPeriods)
	s.runJob(ctx, "notify expiring and traffic", s.services.Workers.NotifyExpiringAndTraffic)
	s.runJob(ctx, "retry failed activations", s.services.Workers.RetryFailedActivations)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-expireTicker.C:
			s.runJob(ctx, "expire subscriptions", s.services.Workers.ExpireSubscriptions)

		case <-deleteTicker.C:
			s.runJob(ctx, "delete old disabled users", s.services.Workers.DeleteOldDisabledUsers)

		case <-retryTicker.C:
			s.runJob(ctx, "retry failed activations", s.services.Workers.RetryFailedActivations)

		case <-syncUsageTicker.C:
			s.runJob(ctx, "sync remnawave usage", s.services.Workers.SyncUsage)

		case <-resetTrafficTicker.C:
			s.runJob(ctx, "reset traffic periods", s.services.Workers.ResetTrafficPeriods)

		case <-notifyTicker.C:
			s.runJob(ctx, "notify expiring and traffic", s.services.Workers.NotifyExpiringAndTraffic)
		}
	}
}

func (s *Scheduler) runJob(ctx context.Context, name string, fn func(context.Context) error) {
	startedAt := time.Now()

	s.log.Info("worker job started", zap.String("job", name))

	if err := fn(ctx); err != nil {
		s.log.Error(
			"worker job failed",
			zap.String("job", name),
			zap.Duration("duration", time.Since(startedAt)),
			zap.Error(err),
		)
		return
	}

	s.log.Info(
		"worker job finished",
		zap.String("job", name),
		zap.Duration("duration", time.Since(startedAt)),
	)
}
