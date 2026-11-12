package app

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"go.uber.org/fx"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	"github.com/suprt/trading/spot-service/internal/infrastructure/config"
)

type seederParams struct {
	fx.In
	LC   fx.Lifecycle
	Cfg  config.Config
	Seed *application.Seed
}

func RunSeeder(p seederParams) {
	if !p.Cfg.Seed.Enabled {
		return
	}

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, sm := range p.Cfg.Seed.Markets {
				roles := make([]domain.Role, len(sm.RequiredRoles))
				for i, r := range sm.RequiredRoles {
					roles[i] = domain.Role(r)
				}
				if err := p.Seed.SeedMarket(ctx, dto.SeedMarketInput{
					Name:          sm.Name,
					RequiredRoles: roles,
					Enabled:       sm.Enabled,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
