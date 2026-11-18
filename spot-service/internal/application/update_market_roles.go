package application

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type UpdateMarketRolesUseCase struct {
	markets outbound.MarketRepository
}

func NewUpdateMarketRolesUseCase(m outbound.MarketRepository) *UpdateMarketRolesUseCase {
	return &UpdateMarketRolesUseCase{markets: m}
}

func (uc *UpdateMarketRolesUseCase) UpdateMarketRoles(ctx context.Context, in dto.UpdateMarketRolesInput) (*dto.UpdateMarketRolesOutput, error) {
	if !auth.IsAdminFromContext(ctx) {
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
