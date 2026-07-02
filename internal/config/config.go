package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"

	repoPool "sakeofher/internal/repository/pool"
)

type Config struct {
	App       AppConfig
	HTTP      HTTPConfig
	Postgres  repoPool.Config
	JWT       JWTConfig
	Telegram  TelegramConfig
	Remnawave RemnawaveConfig
	Tribute   TributeConfig
	CryptoBot CryptoBotConfig
	Worker    WorkerConfig
}

type AppConfig struct {
	Env       string `envconfig:"APP_ENV" default:"local"`
	PublicURL string `envconfig:"APP_PUBLIC_URL" default:"http://localhost:8080"`
}

type HTTPConfig struct {
	Addr string `envconfig:"HTTP_ADDR" default:":8080"`
}

type JWTConfig struct {
	Secret     string        `envconfig:"JWT_SECRET" default:"change_me"`
	AccessTTL  time.Duration `envconfig:"JWT_ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"JWT_REFRESH_TTL" default:"720h"`
}

type TelegramConfig struct {
	BotToken    string  `envconfig:"TELEGRAM_BOT_TOKEN"`
	AdminIDs    []int64 `ignored:"true"`
	RawAdminIDs string  `envconfig:"TELEGRAM_ADMIN_IDS" default:""`
}

type RemnawaveConfig struct {
	BaseURL string        `envconfig:"REMNAWAVE_BASE_URL" default:""`
	Token   string        `envconfig:"REMNAWAVE_API_TOKEN" default:""`
	Timeout time.Duration `envconfig:"REMNAWAVE_TIMEOUT" default:"15s"`
}

type TributeConfig struct {
	APIKey            string        `envconfig:"TRIBUTE_API_KEY" default:""`
	WebhookSecretPath string        `envconfig:"TRIBUTE_WEBHOOK_SECRET_PATH" default:"tribute-secret"`
	Timeout           time.Duration `envconfig:"TRIBUTE_TIMEOUT" default:"15s"`
}

type CryptoBotConfig struct {
	APIToken          string        `envconfig:"CRYPTOBOT_API_TOKEN" default:""`
	WebhookSecretPath string        `envconfig:"CRYPTOBOT_WEBHOOK_SECRET_PATH" default:"cryptobot-secret"`
	Timeout           time.Duration `envconfig:"CRYPTOBOT_TIMEOUT" default:"15s"`
}

type WorkerConfig struct {
	SyncUsageInterval       time.Duration `envconfig:"WORKER_SYNC_USAGE_INTERVAL" default:"1h"`
	ExpireInterval          time.Duration `envconfig:"WORKER_EXPIRE_INTERVAL" default:"1h"`
	DeleteDisabledInterval  time.Duration `envconfig:"WORKER_DELETE_DISABLED_INTERVAL" default:"24h"`
	ResetTrafficInterval    time.Duration `envconfig:"WORKER_RESET_TRAFFIC_INTERVAL" default:"1h"`
	NotifyInterval          time.Duration `envconfig:"WORKER_NOTIFY_INTERVAL" default:"24h"`
	RetryActivationInterval time.Duration `envconfig:"WORKER_RETRY_ACTIVATION_INTERVAL" default:"5m"`
}

func Load() (Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg.App); err != nil {
		return Config{}, fmt.Errorf("process app config: %w", err)
	}
	if err := envconfig.Process("", &cfg.HTTP); err != nil {
		return Config{}, fmt.Errorf("process http config: %w", err)
	}
	pg, err := repoPool.NewConfig()
	if err != nil {
		return Config{}, err
	}
	cfg.Postgres = pg
	if err := envconfig.Process("", &cfg.JWT); err != nil {
		return Config{}, fmt.Errorf("process jwt config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Telegram); err != nil {
		return Config{}, fmt.Errorf("process telegram config: %w", err)
	}
	cfg.Telegram.AdminIDs = parseAdminIDs(cfg.Telegram.RawAdminIDs)
	if err := envconfig.Process("", &cfg.Remnawave); err != nil {
		return Config{}, fmt.Errorf("process remnawave config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Tribute); err != nil {
		return Config{}, fmt.Errorf("process tribute config: %w", err)
	}
	if err := envconfig.Process("", &cfg.CryptoBot); err != nil {
		return Config{}, fmt.Errorf("process cryptobot config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Worker); err != nil {
		return Config{}, fmt.Errorf("process worker config: %w", err)
	}

	return cfg, nil
}

func parseAdminIDs(raw string) []int64 {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]int64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseInt(part, 10, 64)
		if err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}
