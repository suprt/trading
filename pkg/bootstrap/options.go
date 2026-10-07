package bootstrap

import (
	"github.com/suprt/trading/pkg/config"
	"go.uber.org/fx"
)

type Options struct {
	Config  config.Config
	Modules []fx.Option
}
