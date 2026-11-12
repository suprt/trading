package e2e_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/suprt/trading/order-service/internal/app"
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
	orderconfig "github.com/suprt/trading/order-service/internal/infrastructure/config"
	orderv1 "github.com/suprt/trading/proto/order/v1"
	sharedconfig "github.com/suprt/trading/shared/config"
	sharedlogger "github.com/suprt/trading/shared/logger"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

const bufSize = 1024 * 1024

// --- моки внешних сервисов ---

type fakeUserClient struct {
	roles []domain.Role
}

func (f *fakeUserClient) GetUserRoles(_ context.Context, _ domain.UserID) ([]domain.Role, int64, error) {
	return f.roles, 1, nil
}

type fakeSpotClient struct {
	markets map[domain.MarketID]*domain.Market
}

func (f *fakeSpotClient) GetMarket(_ context.Context, id domain.MarketID) (*domain.Market, error) {
	m, ok := f.markets[id]
	if !ok {
		return nil, er.ErrMarketNotFound
	}
	return m, nil
}

// --- setup ---

type testEnv struct {
	client  orderv1.OrderServiceClient
	users   *fakeUserClient
	spots   *fakeSpotClient
	cleanup func()
}

func setupTestServer(t *testing.T, userRoles []domain.Role, markets map[domain.MarketID]*domain.Market) *testEnv {
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

	// 2. Redis
	redisContainer, err := tcredis.Run(ctx, "redis:8-alpine")
	require.NoError(t, err)
	redisURL, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)
	redisOpts, err := redis.ParseURL(redisURL)
	require.NoError(t, err)

	// 3. bufconn
	lis := bufconn.Listen(bufSize)

	// 4. конфиг
	cfg := orderconfig.Config{
		GRPC: orderconfig.GRPC{
			Addr:        "bufconn",
			StopTimeout: 5 * time.Second,
		},
		Postgres: orderconfig.Postgres{
			DSN:             connStr,
			MaxConns:        5,
			MinConns:        1,
			MaxConnLifetime: time.Minute,
			MaxConnIdleTime: time.Minute,
		},
		JWT:         orderconfig.JWT{Secret: "test-secret"},
		Internal:    orderconfig.Internal{Secret: "test-internal-secret"},
		UserService: orderconfig.UserService{Addr: "user-service:50051"},
		SpotService: orderconfig.SpotService{Addr: "spot-service:50052"},
		Level: "debug", Format: "console",
		Default: sharedconfig.RateLimiterRule{Limit: 100, Window: time.Minute},
		Addr:        redisOpts.Addr,
		Username:    "",
		Password:    "",
		DB:          0,
		PingTimeout: 10 * time.Second,

	}

	users := &fakeUserClient{roles: userRoles}
	spots := &fakeSpotClient{markets: markets}

	// 5. fx.App
	newApp := fx.New(
		fx.NopLogger,
		app.Module(cfg),
		fx.Provide(func() *zap.Logger {
			logger, err := sharedlogger.New(sharedlogger.Config{Level: "debug", Format: "console"})
			require.NoError(t, err)
			return logger
		}),
		fx.Decorate(func(_ outbound.UserClient) outbound.UserClient { return users }),
		fx.Decorate(func(_ outbound.SpotClient) outbound.SpotClient { return spots }),
		fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger, interceptors []grpc.UnaryServerInterceptor) *grpc.Server {
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
		}),
	)
	require.NoError(t, newApp.Err())
	require.NoError(t, newApp.Start(ctx))

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		_ = newApp.Stop(ctx)
		_ = conn.Close()
		_ = pgContainer.Terminate(ctx)
		_ = redisContainer.Terminate(ctx)
	}

	return &testEnv{
		client:  orderv1.NewOrderServiceClient(conn),
		users:   users,
		spots:   spots,
		cleanup: cleanup,
	}
}

func authCtx(t *testing.T, userID string) context.Context {
	t.Helper()
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	return metadata.AppendToOutgoingContext(context.Background(), "authorization", "Bearer "+signed)
}