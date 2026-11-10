package config

import "time"

type Log struct {
	Level              string `env:"LOG_LEVEL" envDefault:"info"`
	Format             string `env:"LOG_FORMAT" envDefault:"console"`
	SamplingInitial    int    `env:"LOG_SAMPLING_INITIAL" envDefault:"100"`
	SamplingThereafter int    `env:"LOG_SAMPLING_THEREAFTER" envDefault:"100"`
}

type Tracing struct {
	Enabled  bool   `env:"TRACE_ENABLED" envDefault:"false"`
	Endpoint string `env:"TRACE_ENDPOINT" envDefault:"127.0.0.1:6831"`
}

type Metrics struct {
	Enabled bool   `env:"METRICS_ENABLED" envDefault:"false"`
	Addr    string `env:"METRICS_ADDR" envDefault:":9090"`
}

type Redis struct {
	Addr        string        `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Username    string        `env:"REDIS_USER" envDefault:"redis_user"`
	Password    string        `env:"REDIS_PASSWORD"`
	DB          int           `env:"REDIS_DB" envDefault:"0"`
	PingTimeout time.Duration `env:"REDIS_PING_TIMEOUT" envDefault:"5s"`
}

type JWTSecret struct {
	Secret string `env:"JWT_SECRET" envDefault:"dev-secret"`
}

type RateLimiterRule struct {
	Limit  int           `yaml:"limit"`
	Window time.Duration `yaml:"window"`
}

type RateLimiter struct {
	Default RateLimiterRule            `yaml:"default"`
	Methods map[string]RateLimiterRule `yaml:"methods"`
}
