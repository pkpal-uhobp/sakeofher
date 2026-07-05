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

	lastReconcileAt time.Time
}

func NewScheduler(services *service.Services, cfg config.WorkerConfig, log *zap.Logger) *Scheduler {
	if log == nil {
		log = zap.NewNop()
	}

	return &Scheduler{
		services: services,
		cfg:      cfg,
		log:      log,
	}
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

	// First run.
	//
	// Important order:
	// 1. Pull Remnawave usage/limit/date into site DB.
	// 2. Apply local lifecycle changes.
	// 3. Push the final site state back to Remnawave.
	//
	// Without this order manual edits in Remnawave can be overwritten by reconcile.
	s.runJob(ctx, "sync remnawave usage", s.services.Workers.SyncUsage)
	s.runJob(ctx, "expire subscriptions", s.services.Workers.ExpireSubscriptions)
	s.runJob(ctx, "reset traffic periods", s.services.Workers.ResetTrafficPeriods)
	s.runReconcile(ctx, false)
	s.runJob(ctx, "notify expiring and traffic", s.services.Workers.NotifyExpiringAndTraffic)
	s.runJob(ctx, "retry failed activations", s.services.Workers.RetryFailedActivations)
	s.runJob(ctx, "delete old disabled users", s.services.Workers.DeleteOldDisabledUsers)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-expireTicker.C:
			// Pull Remnawave changes before checking local expiration.
			s.runJob(ctx, "sync remnawave usage", s.services.Workers.SyncUsage)
			s.runJob(ctx, "expire subscriptions", s.services.Workers.ExpireSubscriptions)
			s.runReconcile(ctx, true)

		case <-deleteTicker.C:
			s.runJob(ctx, "delete old disabled users", s.services.Workers.DeleteOldDisabledUsers)

		case <-retryTicker.C:
			s.runJob(ctx, "retry failed activations", s.services.Workers.RetryFailedActivations)

		case <-syncUsageTicker.C:
			s.runJob(ctx, "sync remnawave usage", s.services.Workers.SyncUsage)
			s.runReconcile(ctx, true)

		case <-resetTrafficTicker.C:
			s.runJob(ctx, "reset traffic periods", s.services.Workers.ResetTrafficPeriods)
			s.runReconcile(ctx, true)

		case <-notifyTicker.C:
			s.runJob(ctx, "notify expiring and traffic", s.services.Workers.NotifyExpiringAndTraffic)
		}
	}
}

func (s *Scheduler) runReconcile(ctx context.Context, throttled bool) {
	const minReconcileGap = 10 * time.Second

	if throttled && !s.lastReconcileAt.IsZero() && time.Since(s.lastReconcileAt) < minReconcileGap {
		s.log.Info(
			"worker job skipped",
			zap.String("job", "reconcile remnawave state"),
			zap.String("reason", "recently reconciled"),
			zap.Duration("min_gap", minReconcileGap),
		)
		return
	}

	s.lastReconcileAt = time.Now()
	s.runJob(ctx, "reconcile remnawave state", s.services.Workers.ReconcileRemnaState)
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
