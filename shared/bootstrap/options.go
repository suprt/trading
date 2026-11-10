package bootstrap

import (
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	Addr         string
	Modules      []fx.Option
	StopTimeout  time.Duration
	StartTimeout time.Duration
	Log          *zap.Logger
}
