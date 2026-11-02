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

type UpdateMarketRolesUseCase struct {
	markets outbound.MarketRepository
	users   outbound.UserClient
	ids     outbound.IDGenerator
}

func NewUpdateMarketRolesUseCase(m outbound.MarketRepository, u outbound.UserClient, ids outbound.IDGenerator) *UpdateMarketRolesUseCase {
	return &UpdateMarketRolesUseCase{markets: m, users: u, ids: ids}
}

func (uc *UpdateMarketRolesUseCase) UpdateMarketRoles(ctx context.Context, in dto.UpdateMarketRolesInput) (*dto.UpdateMarketRolesOutput, error) {
	roles, _, err := uc.users.GetUserRoles(ctx, in.CallerID)
	if err != nil {
		return nil, err
	}

	hasAdmin := slices.Contains(roles, domain.RoleAdmin)
	if !hasAdmin {
		return nil, er.ErrForbidden
	}

	m, err := uc.markets.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !m.IsActive() {
		return nil, er.ErrMarketNotFound
	}

	m.RequiredRoles = in.RequiredRoles
	m.UpdatedAt = time.Now()

	if err := uc.markets.UpdateRoles(ctx, m); err != nil {
		return nil, err
	}
	return &dto.UpdateMarketRolesOutput{Market: m}, nil
}
