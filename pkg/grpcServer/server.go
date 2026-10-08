package grpcServer

import (
	"context"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type Params struct {
	fx.In
	LC   fx.Lifecycle
	Addr Addr `name: "grpc_addr"`
}
type Addr string
type Result struct {
	fx.Out
	Server *grpc.Server
}

func New(p Params) Result {
	srv := grpc.NewServer()

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", string(p.Addr))
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
