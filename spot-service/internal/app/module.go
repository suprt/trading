package app

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/suprt/trading/shared/bootstrap"
	sharedconfig "github.com/suprt/trading/shared/config"
	"github.com/suprt/trading/shared/logger"
	"github.com/suprt/trading/shared/zaperror"
	"github.com/suprt/trading/spot-service/internal/adapters/outbound/cached"
	"github.com/suprt/trading/spot-service/internal/adapters/outbound/rediscache"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	spotv1 "github.com/suprt/trading/proto/spot/v1"
	userv1 "github.com/suprt/trading/proto/user/v1"
	"github.com/suprt/trading/shared/db"
	"github.com/suprt/trading/shared/idgen"
	sharedinters "github.com/suprt/trading/shared/interceptors"
	shareredis "github.com/suprt/trading/shared/redis"
	grpchand "github.com/suprt/trading/spot-service/internal/adapters/inbound/grpc"
	spotinters "github.com/suprt/trading/spot-service/internal/adapters/inbound/interceptors"
	spotjwt "github.com/suprt/trading/spot-service/internal/adapters/outbound/jwt"
	"github.com/suprt/trading/spot-service/internal/adapters/outbound/postgres"
	"github.com/suprt/trading/spot-service/internal/adapters/outbound/userclient"
	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/ports/inbound"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
	spotconfig "github.com/suprt/trading/spot-service/internal/infrastructure/config"
	"github.com/suprt/trading/spot-service/internal/infrastructure/migrations"
)

func Module(cfg spotconfig.Config) fx.Option {
	return fx.Module("spot",
		fx.Provide(func() spotconfig.Config { return cfg }),

		// --- ErrorReporter ---
		fx.Provide(zaperror.NewReporter),
		fx.Provide(func(r *zaperror.Reporter) outbound.ErrorReporter { return r }),
		fx.Provide(func(r *zaperror.Reporter) sharedinters.ErrorReporter { return r }),
		fx.Provide(func(r *zaperror.Reporter) shareredis.ErrorReporter { return r }),
		fx.Provide(func(r *zaperror.Reporter) db.ErrorReporter { return r }),

		// --- Postgres ---
		fx.Provide(
			func(cfg spotconfig.Config) db.Config {
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

		// --- Redis ---
		fx.Provide(
			func(cfg spotconfig.Config) shareredis.Config {
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
		fx.Provide(
			func(client *redis.Client, cfg spotconfig.Config) outbound.MarketCache {
				return rediscache.NewMarketCache(client, cfg.Cache.MarketsTTL, cfg.Cache.LockTTL)
			},
		),

		// --- Миграции ---
		fx.Invoke(func(lc fx.Lifecycle, cfg spotconfig.Config) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return migrations.Up(ctx, cfg.Postgres.DSN)
				},
			})
		}),

		// --- Client user-service ---
		fx.Provide(
			func(cfg spotconfig.Config) userv1.UserServiceClient {
				conn, err := grpc.NewClient(
					cfg.UserService.Addr,
					grpc.WithTransportCredentials(insecure.NewCredentials()),
				)
				if err != nil {
					panic(err)
				}
				return userv1.NewUserServiceClient(conn)
			},
			fx.Annotate(
				func(c userv1.UserServiceClient, cfg spotconfig.Config) outbound.UserClient {
					return userclient.NewUserClient(c, cfg.Internal.Secret)
				},
			),
		),

		// --- JWT ---
		fx.Provide(
			func(cfg spotconfig.Config) []byte {
				return []byte(cfg.JWT.Secret)
			},
			spotjwt.NewParser,
			fx.Annotate(spotjwt.NewParser, fx.As(new(outbound.TokenIssuer))),
		),

		// --- RateLimiter ---
		fx.Provide(
			func(cfg spotconfig.Config) sharedinters.RateLimiterConfig {
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
				cfg spotconfig.Config,
			) []grpc.UnaryServerInterceptor {
				return []grpc.UnaryServerInterceptor{
					sharedinters.PanicRecovery(logger),
					sharedinters.XRequestID(),
					sharedinters.Logger(logger),
					sharedinters.UnaryRateLimiter(redisClient, rlCfg, errReporter),
					spotinters.UnaryAuth(tokens, cfg.Auth.Whitelist),
					sharedinters.UnaryValidation(),
				}
			},
		),

		//-- Outbound adapters --
		fx.Provide(
			postgres.NewMarketRepo,
			fx.Annotate(idgen.NewUUIDGen, fx.As(new(outbound.IDGenerator))),
			fx.Annotate(
				func(repo *postgres.MarketRepo, cache outbound.MarketCache, rep outbound.ErrorReporter) outbound.MarketRepository {
					return cached.NewMarketRepo(repo, cache, rep)
				},
			),
		),

		// --- Use cases ---
		fx.Provide(
			application.NewViewMarketsUseCase,
			application.NewGetMarketUseCase,
			application.NewCreateMarketUseCase,
			application.NewUpdateMarketDetailsUseCase,
			application.NewUpdateMarketRolesUseCase,
			application.NewDeleteMarketUseCase,
			application.NewListAllMarketsUseCase,
			application.NewService,
		),

		fx.Provide(fx.Annotate(func(s *application.Service) inbound.SpotService { return s })),

		fx.Provide(grpchand.New),
		fx.Invoke(registerGRPCHandlers),

		fx.Provide(application.NewSeed),
		fx.Invoke(RunSeeder),
	)
}

func registerGRPCHandlers(srv *grpc.Server, h *grpchand.Handler) {
	spotv1.RegisterSpotInstrumentServiceServer(srv, h)
	reflection.Register(srv)
}

func Start() {
	_ = godotenv.Load()

	yamlPath := os.Getenv("CONFIG_YAML_PATH")
	if yamlPath == "" {
		yamlPath = "config.yaml"
	}
	cfg, err := spotconfig.Load(yamlPath)
	if err != nil {
		log.Fatal(err)
	}

	zapLogger, err := logger.New(logger.Config{
		Level:  cfg.Log.Level,
		Format: cfg.Log.Format,
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
		Modules:      []fx.Option{Module(*cfg)},
		StopTimeout:  cfg.GRPC.StopTimeout,
		StartTimeout: cfg.GRPC.StartTimeout,
		Log:          zapLogger,
	}).Run()
	sugar.Info("service stopped")
}
