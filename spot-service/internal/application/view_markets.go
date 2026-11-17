package application

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type ViewMarketsUseCase struct {
	markets outbound.MarketRepository
}

func NewViewMarketsUseCase(markets outbound.MarketRepository) *ViewMarketsUseCase {
	return &ViewMarketsUseCase{markets: markets}
}

func (uc *ViewMarketsUseCase) ViewMarkets(ctx context.Context, in dto.ViewMarketsInput) (*dto.ViewMarketsOutput, error) {
	roles, ok := auth.RolesFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	all, err := uc.markets.List(ctx)
	if err != nil {
		return nil, err
	}

	var result []*domain.Market
	for _, market := range all {
		if market.AvailableFor(roles) {
			result = append(result, market)
		}
	}
	return &dto.ViewMarketsOutput{Markets: result}, nil
}
