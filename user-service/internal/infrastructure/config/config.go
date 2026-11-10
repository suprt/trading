package config

import (
	"time"

	sharedconfig "github.com/suprt/trading/shared/config"
)

type GRPC struct {
	Addr         string        `env:"GRPC_ADDR" envDefault:":50051"`
	StopTimeout  time.Duration `env:"GRPC_STOP_TIMEOUT" envDefault:"15s"`
	StartTimeout time.Duration `env:"GRPC_START_TIMEOUT" envDefault:"30s"`
}

type JWT struct {
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
}

type SeedUser struct {
	Email    string
	Username string
	Password string
	Roles    []string
}

type Seed struct {
	Enabled bool
	Users   []SeedUser
}
type Config struct {
	sharedconfig.Log
	sharedconfig.JWTSecret
	sharedconfig.Redis
	GRPC                     GRPC
	Postgres                 Postgres
	JWT                      JWT
	Crypto                   Crypto
	Cache                    Cache
	Idempotency              Idempotency
	Internal                 Internal
	sharedconfig.RateLimiter `yaml:"rate_limiter"`
	Seed                     Seed `yaml:"seed"`
	Auth                     Auth `yaml:"auth"`
	Sessions                 Sessions
	Timeouts                 Timeouts
}

type Sessions struct {
	CleanupEnabled  bool          `env:"SESSIONS_CLEANUP_ENABLED"  envDefault:"true"`
	CleanupInterval time.Duration `env:"SESSIONS_CLEANUP_INTERVAL" envDefault:"15m"`
}

type Timeouts struct {
	PostgresConnect time.Duration `env:"TIMEOUT_POSTGRES_CONNECT" envDefault:"10s"`
	SessionCleanup  time.Duration `env:"TIMEOUT_SESSION_CLEANUP" envDefault:"20s"`
}

type Idempotency struct {
	TTL time.Duration `env:"Idempotency_TTL" envDefault:"1h"`
}
type Cache struct {
	UserTTL         time.Duration `env:"CACHE_USER_TTL" envDefault:"2h"`
	UserNotFoundTTL time.Duration `env:"CACHE_USER_NOT_FOUND_TTL" envDefault:"30s"`
}
type Auth struct {
	Whitelist map[string]struct{}
}

type Crypto struct {
	HashCost int `env:"CRYPTO_HASH_COST" envDefault:"10"`
}

type Postgres struct {
	DSN             string        `env:"POSTGRES_DSN,required"`
	MaxConns        int32         `env:"POSTGRES_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"POSTGRES_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"POSTGRES_MAX_CONN_LIFETIME" envDefault:"5m"`
	MaxConnIdleTime time.Duration `env:"POSTGRES_MAX_CONN_IDLE_TIME" envDefault:"1m"`
}

type Internal struct {
	Secret string `env:"INTERNAL_SECRET" envDefault:"dev-internal-secret"`
}

func Load(yamlPath string) (*Config, error) {
	return sharedconfig.Load[Config](yamlPath)
}
