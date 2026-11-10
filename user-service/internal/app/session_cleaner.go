package app

import (
	"context"
	"time"

	userconfig "github.com/suprt/trading/user-service/internal/infrastructure/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type SessionCleaner interface {
	CleanupExpired(ctx context.Context) (int64, error)
}
type sessionCleanerParams struct {
	fx.In
	LC      fx.Lifecycle
	Cfg     userconfig.Config
	Cleaner SessionCleaner
	Logger  *zap.Logger
}

func RunSessionCleaner(p sessionCleanerParams) {
	if !p.Cfg.Sessions.CleanupEnabled {
		p.Logger.Info("session cleaner disabled")
		return
	}

	interval := p.Cfg.Sessions.CleanupInterval
	timeout := p.Cfg.Timeouts.SessionCleanup

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						runCtx, cancel := context.WithTimeout(context.Background(), timeout)
						n, err := p.Cleaner.CleanupExpired(runCtx)
						cancel()
						if err != nil {
							p.Logger.Warn("session cleanup failed", zap.Error(err))
							continue
						}
						if n > 0 {
							p.Logger.Info("expired sessions cleaned", zap.Int64("count", n))
						}

					}
				}
			}()
			return nil
		},
	})
}
