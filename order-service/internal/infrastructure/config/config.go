package config

import (
	"time"

	sharedconfig "github.com/suprt/trading/shared/config"
)

type GRPC struct {
	Addr         string        `env:"GRPC_ADDR" envDefault:":50053"`
	StopTimeout  time.Duration `env:"GRPC_STOP_TIMEOUT" envDefault:"15s"`
	StartTimeout time.Duration `env:"GRPC_START_TIMEOUT" envDefault:"30s"`
}

type Postgres struct {
	DSN             string        `env:"POSTGRES_DSN,required"`
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"5m"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"1m"`
}

type JWT struct {
	Secret string `env:"JWT_SECRET" envDefault:"dev-secret"`
}

type Internal struct {
	Secret string `env:"INTERNAL_SECRET" envDefault:"dev-internal-secret"`
}

type UserService struct {
	Addr string `env:"USER_SERVICE_ADDR" envDefault:"localhost:50051"`
}

type SpotService struct {
	Addr string `env:"SPOT_SERVICE_ADDR" envDefault:"localhost:50052"`
}

type Timeouts struct {
	PostgresConnect time.Duration `env:"TIMEOUT_POSTGRES_CONNECT" envDefault:"10s"`
}

type Config struct {
	sharedconfig.Log
	sharedconfig.Redis
	GRPC                     GRPC
	Postgres                 Postgres
	JWT                      JWT
	Internal                 Internal
	UserService              UserService
	SpotService              SpotService
	Timeouts                 Timeouts
	sharedconfig.RateLimiter `yaml:"rate_limiter"`
}

func Load(yamlPath string) (*Config, error) {
	return sharedconfig.Load[Config](yamlPath)
}
