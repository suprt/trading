package grpcServer

import (
	"context"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"github.com/suprt/ITK-proj/pkg/config"
)

type Params struct {
	fx.In
	LC  fx.Lifecycle
	Cfg config.Config
}

type Result struct {
	fx.Out
	Server *grpc.Server
}

func New(p Params) Result {
	srv := grpc.NewServer()

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", p.Cfg.GRPC.Addr)
			if err != nil {
				return err
			}
			go func() { _ = srv.Serve(lis) }()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.GracefulStop()
			return nil
		},
	})

	return Result{Server: srv}
}
