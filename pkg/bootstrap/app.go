package bootstrap

import (
	"github.com/suprt/ITK-proj/pkg/config"
	"github.com/suprt/ITK-proj/pkg/grpcServer"
	"go.uber.org/fx"
)

func NewApp(opts Options) *fx.App {
	return fx.New(
		fx.Provide(func() config.Config { return opts.Config }),
		fx.Provide(grpcServer.New),
		fx.Options(opts.Modules...))
}
