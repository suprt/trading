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

type UpdateMarketDetailsUseCase struct {
	markets outbound.MarketRepository
	users   outbound.UserClient
	ids     outbound.IDGenerator
}

func NewUpdateMarketDetailsUseCase(m outbound.MarketRepository, u outbound.UserClient, ids outbound.IDGenerator) *UpdateMarketDetailsUseCase {
	return &UpdateMarketDetailsUseCase{markets: m, users: u, ids: ids}
}

func (uc *UpdateMarketDetailsUseCase) UpdateMarketDetails(ctx context.Context,
	in dto.UpdateMarketDetailsInput) (*dto.UpdateMarketDetailsOutput, error) {

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

	if in.Name != nil {
		m.Name = *in.Name
	}
	if in.Enabled != nil {
		m.Enabled = *in.Enabled
	}
	m.UpdatedAt = time.Now()

	if err := uc.markets.UpdateDetails(ctx, m); err != nil {
		return nil, err
	}
	return &dto.UpdateMarketDetailsOutput{Market: m}, nil
}
