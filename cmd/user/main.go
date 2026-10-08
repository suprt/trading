package main

import (
	"github.com/suprt/trading/pkg/bootstrap"
	"github.com/suprt/trading/services/user"
	"go.uber.org/fx"
)

func main() {
	bootstrap.NewApp(bootstrap.Options{
		Modules: []fx.Option{user.Module},
	}).Run()
}
