package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/suprt/trading/shared/config"
)

type Config struct {
	config.Redis
}
type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}

func NewClient(cfg Config, reporter ErrorReporter) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.PingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		reporter.Report(ctx, err, map[string]string{
			"op": "redis_newClient_ping",
		})
	}

	return client, nil
}
