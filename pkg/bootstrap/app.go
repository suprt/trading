package bootstrap

import (
	"github.com/suprt/trading/pkg/config"
	"github.com/suprt/trading/pkg/grpcServer"
	"go.uber.org/fx"
)

func NewApp(opts Options) *fx.App {
	return fx.New(
		fx.Provide(func() config.Config { return opts.Config }),
		fx.Provide(grpcServer.New),
		fx.Options(opts.Modules...))
}
