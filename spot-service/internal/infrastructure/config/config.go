package config

import (
	"time"

	sharedconfig "github.com/suprt/trading/shared/config"
)

type Config struct {
	sharedconfig.Log
	sharedconfig.Redis

	UserService UserService
	Internal    Internal
	JWT         JWT
	GRPC        GRPC
	Postgres    Postgres
	Cache       Cache
	
	sharedconfig.RateLimiter `yaml:"rate_limiter"`
	Auth                     Auth `yaml:"auth"`
	Seed                     Seed `yaml:"seed"`
}

type UserService struct {
	Addr string `env:"USER_SERVICE_ADDR" envDefault:"user-service:50051"`
}

type Internal struct {
	Secret string `env:"INTERNAL_SECRET" envDefault:"dev-internal-secret"`
}

type GRPC struct {
	Addr        string        `env:"GRPC_ADDR" envDefault:":50052"`
	StopTimeout time.Duration `env:"GRPC_STOP_TIMEOUT" envDefault:"15s"`
}

type Postgres struct {
	DSN             string        `env:"POSTGRES_DSN,required"`
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"5m"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"1m"`
}

type Cache struct {
	MarketsTTL time.Duration `env:"CACHE_MARKETS_TTL" envDefault:"24h"`
}

type Auth struct {
	Whitelist map[string]struct{}
}

type SeedMarket struct {
	Name          string   `yaml:"name"`
	RequiredRoles []string `yaml:"required_roles"`
	Enabled       bool     `yaml:"enabled"`
}

type Seed struct {
	Enabled bool         `yaml:"enabled"`
	Markets []SeedMarket `yaml:"markets"`
}

type JWT struct {
	Secret string `env:"JWT_SECRET,required"`
}

func Load(yamlPath string) (*Config, error) {
	return sharedconfig.Load[Config](yamlPath)
}
