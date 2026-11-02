package application

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type GetMarketUseCase struct {
	markets outbound.MarketRepository
}

func NewGetMarketUseCase(markets outbound.MarketRepository) *GetMarketUseCase {
	return &GetMarketUseCase{markets: markets}
}

func (uc *GetMarketUseCase) GetMarket(ctx context.Context, in dto.GetMarketInput) (*dto.GetMarketOutput, error) {
	m, err := uc.markets.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if !m.IsActive() {
		return nil, er.ErrMarketNotFound
	}
	return &dto.GetMarketOutput{Market: m}, nil
}
