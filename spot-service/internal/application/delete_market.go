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

type DeleteMarketUseCase struct {
	markets outbound.MarketRepository
	users   outbound.UserClient
	ids     outbound.IDGenerator
}

func NewDeleteMarketUseCase(m outbound.MarketRepository, u outbound.UserClient, ids outbound.IDGenerator) *DeleteMarketUseCase {
	return &DeleteMarketUseCase{markets: m, users: u, ids: ids}
}

func (uc *DeleteMarketUseCase) DeleteMarket(ctx context.Context, in dto.DeleteMarketInput) error {
	roles, _, err := uc.users.GetUserRoles(ctx, in.CallerID)
	if err != nil {
		return err
	}

	hasAdmin := slices.Contains(roles, domain.RoleAdmin)
	if !hasAdmin {
		return er.ErrForbidden
	}

	return uc.markets.SoftDelete(ctx, in.ID, time.Now())
}
