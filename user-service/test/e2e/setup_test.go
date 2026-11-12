package e2e_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	userv1 "github.com/suprt/trading/proto/user/v1"
	sharedconfig "github.com/suprt/trading/shared/config"
	sharedlogger "github.com/suprt/trading/shared/logger"
	"github.com/suprt/trading/user-service/internal/app"
	userconfig "github.com/suprt/trading/user-service/internal/infrastructure/config"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (userv1.UserServiceClient, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. Postgres
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Redis
	redisContainer, err := tcredis.Run(ctx, "redis:8-alpine")
	require.NoError(t, err)

	redisURL, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	redisOpts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	// 2. bufconn
	lis := bufconn.Listen(bufSize)

	// 3. cfg
	cfg := userconfig.Config{
		Secret: "test-secret",
		GRPC: userconfig.GRPC{
			Addr:        "bufconn", // не используется
			StopTimeout: 5 * time.Second,
		},
		Postgres: userconfig.Postgres{
			DSN:             connStr,
			MaxConns:        5,
			MinConns:        1,
			MaxConnLifetime: time.Minute,
		},
		JWT: userconfig.JWT{
			AccessTTL:  15 * time.Minute,
			RefreshTTL: 168 * time.Hour,
		},
		Crypto: userconfig.Crypto{HashCost: 4}, // быстрый bcrypt
		Level:  "debug", Format: "console",
		Auth: userconfig.Auth{
			Whitelist: map[string]struct{}{
				"/user.v1.UserService/Register":         {},
				"/user.v1.UserService/Login":            {},
				"/user.v1.UserService/RefreshToken":     {},
				"/user.v1.UserService/GetPublicProfile": {},
			},
		},
		Seed: userconfig.Seed{
			Enabled: true,
			Users: []userconfig.SeedUser{
				{Email: "admin@test", Username: "admin", Password: "admin12345", Roles: []string{"USER", "ADMIN"}},
			},
		},
		Default: sharedconfig.RateLimiterRule{
			Limit:  100,
			Window: time.Minute,
		},
		Methods: map[string]sharedconfig.RateLimiterRule{
			"/user.v1.UserService/Login":    {Limit: 5, Window: time.Minute}, // ← 5 для теста
			"/user.v1.UserService/Register": {Limit: 5, Window: time.Minute},
		},

		Addr:        redisOpts.Addr,
		Username:    "",
		Password:    "",
		DB:          0,
		PingTimeout: 10 * time.Second,
	}

	// 4. fx.App
	newApp := fx.New(
		fx.NopLogger,
		app.Module(cfg),
		fx.Provide(
			func() *zap.Logger {
				logger, err := sharedlogger.New(sharedlogger.Config{
					Level:  "debug",
					Format: "console",
				})
				require.NoError(t, err)
				return logger
			},
		),
		fx.Provide(
			func(lc fx.Lifecycle, logger *zap.Logger, interceptors []grpc.UnaryServerInterceptor) *grpc.Server {
				srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() { _ = srv.Serve(lis) }()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						srv.GracefulStop()
						return nil
					},
				})
				return srv
			},
		),
	)
	require.NoError(t, newApp.Err())
	require.NoError(t, newApp.Start(ctx))

	// 5. Клиент
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := userv1.NewUserServiceClient(conn)

	cleanup := func() {
		_ = newApp.Stop(ctx)
		_ = conn.Close()
		_ = pgContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
	}
	return client, cleanup
}
