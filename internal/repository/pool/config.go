package pool

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host              string        `envconfig:"HOST" required:"true"`
	Port              string        `envconfig:"PORT" default:"5432"`
	User              string        `envconfig:"USER" required:"true"`
	Password          string        `envconfig:"PASSWORD" required:"true"`
	Database          string        `envconfig:"DB" required:"true"`
	SSLMode           string        `envconfig:"SSL_MODE" default:"disable"`
	MaxConns          int32         `envconfig:"MAX_CONNS" default:"20"`
	MinConns          int32         `envconfig:"MIN_CONNS" default:"2"`
	MaxConnLifetime   time.Duration `envconfig:"MAX_CONN_LIFETIME" default:"1h"`
	MaxConnIdleTime   time.Duration `envconfig:"MAX_CONN_IDLE_TIME" default:"30m"`
	HealthCheckPeriod time.Duration `envconfig:"HEALTH_CHECK_PERIOD" default:"1m"`
	ConnectTimeout    time.Duration `envconfig:"CONNECT_TIMEOUT" default:"5s"`
	QueryTimeout      time.Duration `envconfig:"QUERY_TIMEOUT" default:"5s"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if err := envconfig.Process("POSTGRES", &cfg); err != nil {
		return Config{}, fmt.Errorf("process postgres config: %w", err)
	}
	return cfg, nil
}
