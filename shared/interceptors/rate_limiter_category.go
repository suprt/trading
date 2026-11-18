package interceptors

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	sharedconfig "github.com/suprt/trading/shared/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type CategoryRateLimiterConfig struct {
	sharedconfig.CategoryRateLimiter
}
type AuthInfo interface {
	IsAdmin(ctx context.Context) bool
	Category(ctx context.Context) (string, bool)
}

func UnaryCategoryRateLimiter(
	client *redis.Client,
	cfg CategoryRateLimiterConfig,
	authInfo AuthInfo,
	reporter ErrorReporter,
) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if authInfo.IsAdmin(ctx) {
			return handler(ctx, req)
		}

		cat, ok := authInfo.Category(ctx)
		if !ok {
			cat = "basic"
		}

		limits := cfg.Basic
		if cat == "premium" {
			limits = cfg.Premium
		}

		ip, err := clientIP(ctx)
		if err != nil {
			return handler(ctx, req)
		}

		ipKey := fmt.Sprintf("rl:ip:%s:%s:minute", ip, cat)
		allowed, err := checkLimit(ctx, client, ipKey, int64(limits.PerIPRPM), time.Minute)
		if err != nil {
			reporter.Report(ctx, err, map[string]string{"op": "rate_limit_ip"})
			return handler(ctx, req)
		}
		if !allowed {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		catKey := fmt.Sprintf("rl:cat:%s:second", cat)
		allowed, err = checkLimit(ctx, client, catKey, int64(limits.BudgetRPS), time.Second)
		if err != nil {
			reporter.Report(ctx, err, map[string]string{"op": "rate_limit_cat"})
			return handler(ctx, req)
		}
		if !allowed {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		return handler(ctx, req)
	}
}
func checkLimit(ctx context.Context, client *redis.Client, key string, limit int64, window time.Duration) (bool, error) {
	count, err := client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		if err := client.Expire(ctx, key, window).Err(); err != nil {
			return false, err
		}
	}
	return count <= limit, nil
}

func clientIP(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		//1. x-real-ip
		if vals := md.Get("x-real-ip"); len(vals) > 0 && vals[0] != "" {
			return vals[0], nil
		}
		//2. last x-forwarded-for
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
			parts := strings.Split(vals[0], ",")
			last := strings.TrimSpace(parts[len(parts)-1])
			if last != "" {
				return last, nil
			}
		}
	}
	//3. peer.Addr
	if p, ok := peer.FromContext(ctx); ok {
		if addr, ok := p.Addr.(*net.TCPAddr); ok {
			return addr.IP.String(), nil
		}
	}
	return "", errors.New("no client ip found")
}
