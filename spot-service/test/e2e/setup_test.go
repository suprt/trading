package e2e_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	sharedconfig "github.com/suprt/trading/shared/config"
	"github.com/suprt/trading/spot-service/internal/app"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	spotv1 "github.com/suprt/trading/proto/spot/v1"
	sharedlogger "github.com/suprt/trading/shared/logger"
	spotconfig "github.com/suprt/trading/spot-service/internal/infrastructure/config"
)

const bufSize = 1024 * 1024

// Мок клиента к user-service (чтобы не поднимать user-service).
type fakeUserClient struct {
	roles []string
}

func (f *fakeUserClient) GetUserRoles(_ context.Context, _ domain.UserID) ([]domain.Role, int, error) {
	roles := make([]domain.Role, len(f.roles))
	for i, r := range f.roles {
		roles[i] = domain.Role(r)
	}
	return roles, 1, nil
}

type fakeTokenIssuer struct{}

func (f *fakeTokenIssuer) ParseAccess(_ domain.AccessToken) (domain.UserID, []domain.Role, int64, error) {
	return "fake-user-id", nil, 0, nil
}

func setupTestServer(t *testing.T, callerRoles []string) (spotv1.SpotInstrumentServiceClient, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. Postgres
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
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
	cfg := spotconfig.Config{
		GRPC: spotconfig.GRPC{
			Addr:        "bufconn",
			StopTimeout: 5 * time.Second,
		},
		Postgres: spotconfig.Postgres{
			DSN:             connStr,
			MaxConns:        5,
			MinConns:        1,
			MaxConnLifetime: time.Minute,
			MaxConnIdleTime: time.Minute,
		},
		Addr:     redisOpts.Addr,
		Username: "",
		Password: "",
		DB:       0,
		JWT: spotconfig.JWT{
			Secret: "test-secret",
		},
		Internal: spotconfig.Internal{
			Secret: "test-internal-secret",
		},
		UserService: spotconfig.UserService{
			Addr: "user-service:50051", // не используется (замокан)
		},
		Cache: spotconfig.Cache{
			MarketsTTL: time.Hour,
		},
		Level:  "debug",
		Format: "console",
		Auth: spotconfig.Auth{
			Whitelist: map[string]struct{}{
				"/spot.v1.SpotInstrumentService/ViewMarkets": {},
				"/spot.v1.SpotInstrumentService/GetMarket":   {},
			},
		},
		Seed: spotconfig.Seed{
			Enabled: true,
			Markets: []spotconfig.SeedMarket{
				{Name: "BTC/USDT", RequiredRoles: []string{}, Enabled: true},
				{Name: "SOL/USDT", RequiredRoles: []string{"TRADER"}, Enabled: true},
			},
		},
		Default: sharedconfig.RateLimiterRule{
			Limit:  100,
			Window: time.Minute,
		},
	}

	// 4. fx.App
	newApp := fx.New(
		fx.NopLogger,
		app.Module(cfg),
		fx.Provide(
			func() *zap.Logger {
				logger, err := sharedlogger.New(sharedlogger.Config{Level: "debug", Format: "console"})
				require.NoError(t, err)
				return logger
			},
		),
		fx.Decorate(
			func(_ outbound.UserClient) outbound.UserClient {
				return &fakeUserClient{roles: callerRoles}
			},
		),
		// Мокаем TokenIssuer
		fx.Decorate(
			func(_ outbound.TokenIssuer) outbound.TokenIssuer {
				return &fakeTokenIssuer{}
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

	client := spotv1.NewSpotInstrumentServiceClient(conn)

	cleanup := func() {
		_ = newApp.Stop(ctx)
		_ = conn.Close()
		_ = pgContainer.Terminate(ctx)
	}
	return client, cleanup
}
