package application

import (
	"context"
	"slices"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type ListAllMarketsUseCase struct {
	markets outbound.MarketRepository
	users   outbound.UserClient
}

func NewListAllMarketsUseCase(markets outbound.MarketRepository, users outbound.UserClient) *ListAllMarketsUseCase {
	return &ListAllMarketsUseCase{markets: markets, users: users}
}

func (uc *ListAllMarketsUseCase) ListAllMarkets(ctx context.Context, in dto.ListAllMarketsInput) (*dto.ListAllMarketsOutput, error) {

	roles, _, err := uc.users.GetUserRoles(ctx, in.CallerID)
	if err != nil {
		return nil, err
	}

	hasAdmin := slices.Contains(roles, domain.RoleAdmin)
	if !hasAdmin {
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
