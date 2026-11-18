package application

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type CreateMarketUseCase struct {
	markets outbound.MarketRepository
	ids     outbound.IDGenerator
}

func NewCreateMarketUseCase(m outbound.MarketRepository, ids outbound.IDGenerator) *CreateMarketUseCase {
	return &CreateMarketUseCase{markets: m, ids: ids}
}

func (uc *CreateMarketUseCase) CreateMarket(ctx context.Context, in dto.CreateMarketInput) (*dto.CreateMarketOutput, error) {
	if !auth.IsAdminFromContext(ctx) {
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
