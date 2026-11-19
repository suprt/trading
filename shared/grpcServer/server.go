package grpcServer

import (
	"context"
	"errors"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Params struct {
	fx.In
	LC           fx.Lifecycle
	Shutdowner   fx.Shutdowner
	Addr         Addr
	Logger       *zap.Logger
	Interceptors []grpc.UnaryServerInterceptor
}
type Addr string
type Result struct {
	fx.Out
	Server *grpc.Server
}

func New(p Params) Result {
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(p.Interceptors...),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", string(p.Addr))
			if err != nil {
				return err
			}
			p.Logger.Info("grpc server listening", zap.String("addr", string(p.Addr)))
			go func() {
				err := srv.Serve(lis)

				if err != nil &&
					!errors.Is(err, grpc.ErrServerStopped) &&
					!errors.Is(err, net.ErrClosed) {
					p.Logger.Error("grpc serve failed", zap.Error(err))
					if shutdownErr := p.Shutdowner.Shutdown(); shutdownErr != nil {
						p.Logger.Error("shutdown failed", zap.Error(shutdownErr))
					}
				}

			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			done := make(chan struct{})
			go func() {
				srv.GracefulStop()
				close(done)
			}()
			select {
			case <-done:
				p.Logger.Info("grpc server stopped gracefully")
				return nil
			case <-ctx.Done():
				srv.Stop()
				p.Logger.Warn("grpc server stopped by timeout")
				return ctx.Err()
			}

		},
	})

	return Result{Server: srv}
}
