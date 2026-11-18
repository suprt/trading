package interceptors_test

import (
	"context"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	sharedconfig "github.com/suprt/trading/shared/config"
	sharedinters "github.com/suprt/trading/shared/interceptors"
)

type fakeAuthInfo struct {
	isAdmin  bool
	category string
}

func (f *fakeAuthInfo) IsAdmin(_ context.Context) bool            { return f.isAdmin }
func (f *fakeAuthInfo) Category(_ context.Context) (string, bool) { return f.category, true }

type noopReporter struct{}

func (n *noopReporter) Report(_ context.Context, _ error, _ map[string]string) {}

func testCfg() sharedinters.CategoryRateLimiterConfig {
	return sharedinters.CategoryRateLimiterConfig{
		Premium: sharedconfig.CategoryLimits{PerIPRPM: 5, BudgetRPS: 1000},
		Basic:   sharedconfig.CategoryLimits{PerIPRPM: 2, BudgetRPS: 1000},
	}
}

func newLimiter(t *testing.T, cfg sharedinters.CategoryRateLimiterConfig, ai sharedinters.AuthInfo) grpc.UnaryServerInterceptor {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return sharedinters.UnaryCategoryRateLimiter(client, cfg, ai, &noopReporter{})
}

func ctxWithIP(ip string) context.Context {
	p := &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP(ip)}}
	return peer.NewContext(context.Background(), p)
}

func handlerOK(_ context.Context, _ any) (any, error) { return nil, nil }

func TestCategoryRateLimiter_AdminSkip(t *testing.T) {
	limiter := newLimiter(t, testCfg(), &fakeAuthInfo{isAdmin: true})
	for i := 0; i < 100; i++ {
		_, err := limiter(ctxWithIP("1.2.3.4"), nil, &grpc.UnaryServerInfo{}, handlerOK)
		require.NoError(t, err)
	}
}

func TestCategoryRateLimiter_BasicLimit(t *testing.T) {
	limiter := newLimiter(t, testCfg(), &fakeAuthInfo{category: "basic"})
	ctx := ctxWithIP("1.2.3.4")

	for i := 0; i < 2; i++ {
		_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
		require.NoError(t, err)
	}
	_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.Error(t, err)
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	require.Equal(t, "too many requests", status.Convert(err).Message())
}

func TestCategoryRateLimiter_PremiumLimit(t *testing.T) {
	limiter := newLimiter(t, testCfg(), &fakeAuthInfo{category: "premium"})
	ctx := ctxWithIP("1.2.3.4")

	for i := 0; i < 5; i++ {
		_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
		require.NoError(t, err)
	}
	_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.Error(t, err)
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	require.Equal(t, "too many requests", status.Convert(err).Message())
}

func TestCategoryRateLimiter_DifferentIPs(t *testing.T) {
	limiter := newLimiter(t, testCfg(), &fakeAuthInfo{category: "basic"})

	for i := 0; i < 2; i++ {
		_, err := limiter(ctxWithIP("1.2.3.4"), nil, &grpc.UnaryServerInfo{}, handlerOK)
		require.NoError(t, err)
	}
	_, err := limiter(ctxWithIP("5.6.7.8"), nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.NoError(t, err)
}

func TestCategoryRateLimiter_CategoryBudget(t *testing.T) {
	cfg := testCfg()
	cfg.Basic.BudgetRPS = 2
	limiter := newLimiter(t, cfg, &fakeAuthInfo{category: "basic"})

	_, err := limiter(ctxWithIP("1.1.1.1"), nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.NoError(t, err)
	_, err = limiter(ctxWithIP("2.2.2.2"), nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.NoError(t, err)
	_, err = limiter(ctxWithIP("3.3.3.3"), nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.Error(t, err)
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	require.Equal(t, "too many requests", status.Convert(err).Message())
}

func TestCategoryRateLimiter_XRealIP(t *testing.T) {
	limiter := newLimiter(t, testCfg(), &fakeAuthInfo{category: "basic"})
	md := metadata.Pairs("x-real-ip", "9.9.9.9")
	ctx := metadata.NewIncomingContext(context.Background(), md)

	for i := 0; i < 2; i++ {
		_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
		require.NoError(t, err)
	}
	_, err := limiter(ctx, nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.Error(t, err)
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
}

func TestCategoryRateLimiter_FailOpen(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	limiter := sharedinters.UnaryCategoryRateLimiter(client, testCfg(), &fakeAuthInfo{category: "basic"}, &noopReporter{})

	_, err := limiter(ctxWithIP("1.2.3.4"), nil, &grpc.UnaryServerInfo{}, handlerOK)
	require.NoError(t, err)
}
