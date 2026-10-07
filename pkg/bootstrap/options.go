package bootstrap

import (
	"github.com/suprt/ITK-proj/pkg/config"
	"go.uber.org/fx"
)

type Options struct {
	Config  config.Config
	Modules []fx.Option
}
