package app

import (
	"context"
	"errors"
	"net/http"

	"go.uber.org/zap"

	httptransport "sakeofher/internal/transport/http"
)

func RunAPI(ctx context.Context) error {
	c, err := NewContainer(ctx)
	if err != nil {
		return err
	}
	defer c.Close()

	router := httptransport.NewRouter(c.Services, c.Config.App.SubscriptionPathSecret, c.Config.Telegram.OAuthSuccessRedirectURL, c.Config.JWT.Secret, c.Log)
	srv := &http.Server{Addr: c.Config.HTTP.Addr, Handler: router}

	c.Log.Info("api started", zap.String("addr", c.Config.HTTP.Addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
