package application

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type ListAllMarketsUseCase struct {
	markets outbound.MarketRepository
}

func NewListAllMarketsUseCase(markets outbound.MarketRepository) *ListAllMarketsUseCase {
	return &ListAllMarketsUseCase{markets: markets}
}

func (uc *ListAllMarketsUseCase) ListAllMarkets(ctx context.Context, in dto.ListAllMarketsInput) (*dto.ListAllMarketsOutput, error) {

	if !auth.IsAdminFromContext(ctx) {
		return nil, er.ErrForbidden
	}

	all, err := uc.markets.List(ctx)
	if err != nil {
		return nil, err
	}

	var result []*domain.Market
	for _, m := range all {
		if !in.IncludeDeleted && m.DeletedAt != nil {
			continue
		}
		if !in.IncludeDisabled && !m.Enabled {
			continue
		}
		result = append(result, m)
	}

	return &dto.ListAllMarketsOutput{Markets: result}, nil
}
