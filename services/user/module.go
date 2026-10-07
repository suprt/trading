package user

import (
	"github.com/suprt/ITK-proj/pkg/config"
	userv1 "github.com/suprt/ITK-proj/pkg/proto/user/v1"
	"github.com/suprt/ITK-proj/services/user/internal/handler"
	"github.com/suprt/ITK-proj/services/user/internal/infra"
	"github.com/suprt/ITK-proj/services/user/internal/repository/inmemory"
	"github.com/suprt/ITK-proj/services/user/internal/service"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var Module = fx.Module("user",
	fx.Provide(
		func(cfg config.Config) infra.JWTConfig {
			return infra.JWTConfig{
				Secret: cfg.JWT.Secret,
				TTL:    cfg.JWT.AccessTTL,
			}
		},
		func(cfg config.Config) service.RegisterConfig {
			return service.RegisterConfig{
				RefreshTTL: cfg.JWT.RefreshTTL,
			}
		},
	),
	fx.Provide(
		fx.Annotate(inmemory.NewUserRepo, fx.As(new(service.UserRepository))),
		fx.Annotate(inmemory.NewCredentialsRepo, fx.As(new(service.CredentialsRepository))),
		fx.Annotate(inmemory.NewSessionRepo, fx.As(new(service.SessionRepository))),
		fx.Annotate(infra.NewBcryptHasher, fx.As(new(service.PasswordHasher))),
		fx.Annotate(infra.NewUUIDGen, fx.As(new(service.IDGenerator))),
		fx.Annotate(infra.NewJWTIssuer, fx.As(new(service.TokenIssuer))),
	),

	fx.Provide(
		fx.Annotate(
			service.NewRegister,
			fx.As(new(handler.RegisterService)),
		),
	),
	fx.Provide(handler.New),

	fx.Invoke(registerGRPCHandlers),
)

func registerGRPCHandlers(srv *grpc.Server, h *handler.Handler) {
	userv1.RegisterUserServiceServer(srv, h)
	reflection.Register(srv)
}
