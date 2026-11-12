package app

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	sharedconfig "github.com/suprt/trading/shared/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	grpchand "github.com/suprt/trading/order-service/internal/adapters/inbound/grpc"
	orderinters "github.com/suprt/trading/order-service/internal/adapters/inbound/interceptors"
	orderjwt "github.com/suprt/trading/order-service/internal/adapters/outbound/jwt"
	"github.com/suprt/trading/order-service/internal/adapters/outbound/postgres"
	"github.com/suprt/trading/order-service/internal/adapters/outbound/spotclient"
	"github.com/suprt/trading/order-service/internal/adapters/outbound/userclient"
	"github.com/suprt/trading/order-service/internal/application"
	"github.com/suprt/trading/order-service/internal/core/ports/inbound"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
	orderconfig "github.com/suprt/trading/order-service/internal/infrastructure/config"
	"github.com/suprt/trading/order-service/internal/infrastructure/migrations"
	orderv1 "github.com/suprt/trading/proto/order/v1"
	spotv1 "github.com/suprt/trading/proto/spot/v1"
	userv1 "github.com/suprt/trading/proto/user/v1"
	"github.com/suprt/trading/shared/bootstrap"
	"github.com/suprt/trading/shared/db"
	"github.com/suprt/trading/shared/idgen"
	sharedinters "github.com/suprt/trading/shared/interceptors"
	"github.com/suprt/trading/shared/logger"
	shareredis "github.com/suprt/trading/shared/redis"
	"github.com/suprt/trading/shared/zaperror"
)

func Module(cfg orderconfig.Config) fx.Option {
	return fx.Module("order",
		fx.Provide(func() orderconfig.Config { return cfg }),

		// --- ErrorReporter ---
		fx.Provide(zaperror.NewReporter),
		fx.Provide(func(r *zaperror.Reporter) outbound.ErrorReporter { return r }),
		fx.Provide(func(r *zaperror.Reporter) sharedinters.ErrorReporter { return r }),
		fx.Provide(func(r *zaperror.Reporter) shareredis.ErrorReporter { return r }),
		// --- Postgres ---
		fx.Provide(
			func(cfg orderconfig.Config) db.Config {
				return db.Config{
					DSN:             cfg.Postgres.DSN,
					MaxConns:        cfg.Postgres.MaxConns,
					MinConns:        cfg.Postgres.MinConns,
					MaxConnLifetime: cfg.Postgres.MaxConnLifetime,
					MaxConnIdleTime: cfg.Postgres.MaxConnIdleTime,
					ConnectTimeout:  cfg.Timeouts.PostgresConnect,
				}
			},
			db.NewPostgres,
			fx.Annotate(db.NewUoW, fx.As(new(outbound.UnitOfWork))),
		),

		// --- Redis (только для rate limiter) ---
		fx.Provide(
			func(cfg orderconfig.Config) shareredis.Config {
				return shareredis.Config{
					Addr:        cfg.Redis.Addr,
					Username:    cfg.Redis.Username,
					Password:    cfg.Redis.Password,
					DB:          cfg.Redis.DB,
					PingTimeout: cfg.Redis.PingTimeout,
				}
			},
			shareredis.NewClient,
		),

		// --- Миграции ---
		fx.Invoke(func(lc fx.Lifecycle, cfg orderconfig.Config) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return migrations.Up(ctx, cfg.Postgres.DSN)
				},
			})
		}),

		// --- JWT ---
		fx.Provide(
			func(cfg orderconfig.Config) []byte {
				return []byte(cfg.JWT.Secret)
			},
			orderjwt.NewParser,
			fx.Annotate(orderjwt.NewParser, fx.As(new(outbound.TokenIssuer))),
		),

		// --- RateLimiter config ---
		fx.Provide(
			func(cfg orderconfig.Config) sharedinters.RateLimiterConfig {
				methods := make(map[string]sharedconfig.RateLimiterRule, len(cfg.RateLimiter.Methods))
				for k, v := range cfg.RateLimiter.Methods {
					methods[k] = sharedconfig.RateLimiterRule{
						Limit:  v.Limit,
						Window: v.Window,
					}
				}
				return sharedinters.RateLimiterConfig{
					Default: sharedconfig.RateLimiterRule{
						Limit:  cfg.RateLimiter.Default.Limit,
						Window: cfg.RateLimiter.Default.Window,
					},
					Methods: methods,
				}
			},
		),

		// --- Interceptors ---
		fx.Provide(
			func(
				logger *zap.Logger,
				tokens outbound.TokenIssuer,
				redisClient *redis.Client,
				rlCfg sharedinters.RateLimiterConfig,
				errReporter outbound.ErrorReporter,
			) []grpc.UnaryServerInterceptor {
				return []grpc.UnaryServerInterceptor{
					sharedinters.PanicRecovery(logger),
					sharedinters.XRequestID(),
					sharedinters.Logger(logger),
					sharedinters.UnaryRateLimiter(redisClient, rlCfg, errReporter),
					orderinters.UnaryAuth(tokens),
				}
			},
		),

		// --- Outbound clients ---
		fx.Provide(
			func(cfg orderconfig.Config) userv1.UserServiceClient {
				conn, err := grpc.NewClient(
					cfg.UserService.Addr,
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
				if err != nil {
					panic(err)
				}
				return userv1.NewUserServiceClient(conn)
			},
			func(cfg orderconfig.Config) spotv1.SpotInstrumentServiceClient {
				conn, err := grpc.NewClient(
					cfg.SpotService.Addr,
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
				if err != nil {
					panic(err)
				}
				return spotv1.NewSpotInstrumentServiceClient(conn)
			},
			fx.Annotate(
				func(c userv1.UserServiceClient, cfg orderconfig.Config) outbound.UserClient {
					return userclient.NewUserClient(c, cfg.Internal.Secret)
				},
			),
			fx.Annotate(
				func(c spotv1.SpotInstrumentServiceClient) outbound.SpotClient {
					return spotclient.NewSpotClient(c)
				},
			),
		),

		// --- Outbound adapters ---
		fx.Provide(
			fx.Annotate(postgres.NewOrderRepo, fx.As(new(outbound.OrderRepository))),
			fx.Annotate(idgen.NewUUIDGen, fx.As(new(outbound.IDGenerator))),
		),

		// --- Use cases ---
		fx.Provide(
			application.NewCreateOrderUseCase,
			application.NewGetOrderUseCase,
			application.NewGetOrderStatusUseCase,
			application.NewAdminGetOrderUseCase,
			application.NewAdminGetOrderStatusUseCase,
			application.NewCancelOrderUseCase,
			application.NewAdminCancelOrderUseCase,
			application.NewListOrdersUseCase,
			application.NewAdminListOrdersUseCase,
			application.NewService,
		),

		fx.Provide(fx.Annotate(func(s *application.Service) inbound.OrderService { return s })),

		fx.Provide(grpchand.New),
		fx.Invoke(registerGRPCHandlers),
	)
}

func registerGRPCHandlers(srv *grpc.Server, h *grpchand.Handler) {
	orderv1.RegisterOrderServiceServer(srv, h)
	reflection.Register(srv)
}

func Start() {
	_ = godotenv.Load()

	yamlPath := os.Getenv("CONFIG_YAML_PATH")
	if yamlPath == "" {
		yamlPath = "config.yaml"
	}
	cfg, err := orderconfig.Load(yamlPath)
	if err != nil {
		log.Fatal(err)
	}

	zapLogger, err := logger.New(logger.Config{
		Level:              cfg.Log.Level,
		Format:             cfg.Log.Format,
		SamplingInitial:    cfg.Log.SamplingInitial,
		SamplingThereafter: cfg.Log.SamplingThereafter,
	})
	if err != nil {
		log.Fatal(err)
	}
	sugar := zapLogger.Sugar()
	sugar.Infow("service starting",
		"name", "spot-service",
		"addr", cfg.GRPC.Addr,
		"log_level", cfg.Log.Level,
		"log_format", cfg.Log.Format,
	)
	bootstrap.NewApp(bootstrap.Options{
		Addr:         cfg.GRPC.Addr,
		StopTimeout:  cfg.GRPC.StopTimeout,
		StartTimeout: cfg.GRPC.StartTimeout,
		Log:          zapLogger,
		Modules:      []fx.Option{Module(*cfg)},
	}).Run()
	sugar.Info("service stopped")
}
