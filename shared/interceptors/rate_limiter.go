package interceptors

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/suprt/trading/shared/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimiterRule struct {
	config.RateLimiterRule
}

type RateLimiterConfig struct {
	config.RateLimiter
}

type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}

func UnaryRateLimiter(client *redis.Client, cfg RateLimiterConfig, reporter ErrorReporter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any,
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		rule, ok := cfg.Methods[info.FullMethod]
		if !ok {
			rule = cfg.Default
		}
		if rule.Limit <= 0 {
			return handler(ctx, req)
		}

		ip, err := clientIP(ctx)
		if err != nil {
			return handler(ctx, req) // не блокируем если не получилось определить IP
		}

		key := fmt.Sprintf("ratelimit:%s:%s", info.FullMethod, ip)
		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			if reporter != nil {
				reporter.Report(ctx, err, map[string]string{
					"op":     "rate_limit_incr",
					"method": info.FullMethod,
					"key":    key,
				})
			}
			return handler(ctx, req)
		}
		if count == 1 {
			err = client.Expire(ctx, key, rule.Window).Err()
			if err != nil {
				if reporter != nil {
					reporter.Report(ctx, err, map[string]string{
						"op":     "rate_limit_expire",
						"method": info.FullMethod,
						"key":    key,
					})
				}
			}
		}

		if count > int64(rule.Limit) {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}
		return handler(ctx, req)
	}
}
