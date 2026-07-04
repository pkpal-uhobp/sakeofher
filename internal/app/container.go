package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"sakeofher/internal/config"
	"sakeofher/internal/gateway"
	"sakeofher/internal/gateway/cryptobot"
	"sakeofher/internal/gateway/remnawave"
	"sakeofher/internal/gateway/telegram"
	"sakeofher/internal/gateway/tribute"
	"sakeofher/internal/platform/logger"
	"sakeofher/internal/repository"
	"sakeofher/internal/repository/pool"
	"sakeofher/internal/service"
)

type Container struct {
	Config       config.Config
	Log          *zap.Logger
	Repositories *repository.Repositories
	Gateways     gateway.Gateways
	Services     *service.Services
	DB           *pool.ConnectionPool
}

func NewContainer(ctx context.Context) (*Container, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log, err := logger.New(cfg.App.Env)
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}

	db, err := pool.NewConnectionPool(ctx, cfg.Postgres)
	if err != nil {
		return nil, err
	}

	repos := repository.NewRepositories(db)

	gates := gateway.Gateways{
		Remnawave: remnawave.NewClient(cfg.Remnawave.BaseURL, cfg.Remnawave.Token, cfg.Remnawave.Timeout),
		Tribute:   tribute.NewClient(cfg.Tribute.APIKey, cfg.Tribute.Timeout),
		CryptoBot: cryptobot.NewClient(cfg.CryptoBot.APIToken, cfg.CryptoBot.Timeout),
		Telegram:  telegram.NewNotifier(cfg.Telegram.BotToken, log),
	}

	services := service.NewServices(
		repos,
		gates,
		cfg.Telegram.BotUsername,
		cfg.App.PublicURL,
		cfg.App.SubscriptionPathSecret,
		cfg.Admin.Username,
		cfg.Admin.Password,
		cfg.JWT.Secret,
		cfg.JWT.AccessTTL,
		log,
	)

	return &Container{Config: cfg, Log: log, Repositories: repos, Gateways: gates, Services: services, DB: db}, nil
}

func (c *Container) Close() {
	if c.DB != nil {
		c.DB.Close()
	}

	if c.Log != nil {
		_ = c.Log.Sync()
	}
}
