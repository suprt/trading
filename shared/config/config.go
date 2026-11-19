package config

import "time"

type Log struct {
	Level              string `env:"LOG_LEVEL" envDefault:"info"`
	Format             string `env:"LOG_FORMAT" envDefault:"console"`
	SamplingInitial    int    `env:"LOG_SAMPLING_INITIAL" envDefault:"100"`
	SamplingThereafter int    `env:"LOG_SAMPLING_THEREAFTER" envDefault:"100"`
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

type CategoryLimits struct {
	PerIPRPM  int `yaml:"per_ip_rpm"`
	BudgetRPS int `yaml:"budget_rps"`
}

type CategoryRateLimiter struct {
	Premium CategoryLimits `yaml:"premium"`
	Basic   CategoryLimits `yaml:"basic"`
}

type Tracing struct {
	Enabled     bool    `env:"TRACING_ENABLED" envDefault:"false"`
	Endpoint    string  `env:"TRACING_ENDPOINT" envDefault:"otel-collector:4317"`
	ServiceName string  `env:"TRACING_SERVICE_NAME"`
	SampleRate  float64 `env:"TRACING_SAMPLE_RATE" envDefault:"1.0"`
}
