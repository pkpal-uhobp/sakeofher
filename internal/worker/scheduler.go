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
	ticker := time.NewTicker(s.cfg.ExpireInterval)
	defer ticker.Stop()

	s.log.Info("worker scheduler started")
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.services.Workers.ExpireSubscriptions(ctx); err != nil {
				s.log.Error("expire subscriptions failed", zap.Error(err))
			}
		}
	}
}
