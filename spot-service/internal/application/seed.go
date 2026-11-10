package application

import (
	"context"
	"errors"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type Seed struct {
	markets outbound.MarketRepository
	ids     outbound.IDGenerator
}

func NewSeed(markets outbound.MarketRepository, ids outbound.IDGenerator) *Seed {
	return &Seed{markets: markets, ids: ids}
}

func (uc *Seed) SeedMarket(ctx context.Context, in dto.SeedMarketInput) error {
	now := time.Now()
	m := &domain.Market{
		ID:            domain.MarketID(uc.ids.New()),
		Name:          in.Name,
		RequiredRoles: in.RequiredRoles,
		Enabled:       in.Enabled,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := uc.markets.Create(ctx, m); err != nil {
		if errors.Is(err, er.ErrMarketNameTaken) {
			return nil
		}
		return err
	}
	return nil
} //TODO: убрать GetByName, Create если ошибка наличия маркета игнор, остальные пробрасываем
