package app

import (
	"context"
	workertransport "sakeofher/internal/worker"
)

func RunWorker(ctx context.Context) error {
	c, err := NewContainer(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	scheduler := workertransport.NewScheduler(c.Services, c.Config.Worker, c.Log)
	return scheduler.Run(ctx)
}
