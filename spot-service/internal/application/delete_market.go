package application

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type DeleteMarketUseCase struct {
	markets outbound.MarketRepository
	ids     outbound.IDGenerator
}

func NewDeleteMarketUseCase(m outbound.MarketRepository) *DeleteMarketUseCase {
	return &DeleteMarketUseCase{markets: m}
}

func (uc *DeleteMarketUseCase) DeleteMarket(ctx context.Context, in dto.DeleteMarketInput) error {
	if !auth.IsAdminFromContext(ctx) {
		return er.ErrForbidden
	}

	return uc.markets.SoftDelete(ctx, in.ID, time.Now())
}
