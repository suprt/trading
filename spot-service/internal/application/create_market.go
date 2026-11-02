package application

import (
	"context"
	"slices"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type CreateMarketUseCase struct {
	markets outbound.MarketRepository
	users   outbound.UserClient
	ids     outbound.IDGenerator
}

func NewCreateMarketUseCase(m outbound.MarketRepository, u outbound.UserClient, ids outbound.IDGenerator) *CreateMarketUseCase {
	return &CreateMarketUseCase{markets: m, users: u, ids: ids}
}

func (uc *CreateMarketUseCase) CreateMarket(ctx context.Context, in dto.CreateMarketInput) (*dto.CreateMarketOutput, error) {
	roles, _, err := uc.users.GetUserRoles(ctx, in.CallerID)
	if err != nil {
		return nil, err
	}

	hasAdmin := slices.Contains(roles, domain.RoleAdmin)
	if !hasAdmin {
		return nil, er.ErrForbidden
	}

	now := time.Now()
	m := &domain.Market{
		ID:            domain.MarketID(uc.ids.New()),
		Name:          in.Name,
		RequiredRoles: in.RequiredRoles,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := uc.markets.Create(ctx, m); err != nil {
		return nil, err
	}
	return &dto.CreateMarketOutput{Market: m}, nil
}
