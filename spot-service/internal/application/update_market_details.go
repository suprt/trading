package application

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type UpdateMarketDetailsUseCase struct {
	markets outbound.MarketRepository
}

func NewUpdateMarketDetailsUseCase(m outbound.MarketRepository) *UpdateMarketDetailsUseCase {
	return &UpdateMarketDetailsUseCase{markets: m}
}

func (uc *UpdateMarketDetailsUseCase) UpdateMarketDetails(ctx context.Context,
	in dto.UpdateMarketDetailsInput) (*dto.UpdateMarketDetailsOutput, error) {

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
