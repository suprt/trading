package main

import (
	"github.com/suprt/ITK-proj/pkg/bootstrap"
	"github.com/suprt/ITK-proj/pkg/config"
	"github.com/suprt/ITK-proj/services/user"
	"go.uber.org/fx"
)

func main() {
	bootstrap.NewApp(bootstrap.Options{
		Config:  config.Default(),
		Modules: []fx.Option{user.Module},
	}).Run()
}
