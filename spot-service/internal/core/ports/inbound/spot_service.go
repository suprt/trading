package inbound

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/dto"
)

type SpotService interface {
	ViewMarkets(ctx context.Context, in dto.ViewMarketsInput) (*dto.ViewMarketsOutput, error)
	GetMarket(ctx context.Context, in dto.GetMarketInput) (*dto.GetMarketOutput, error)
	CreateMarket(ctx context.Context, in dto.CreateMarketInput) (*dto.CreateMarketOutput, error)
	DeleteMarket(ctx context.Context, in dto.DeleteMarketInput) error
	ListAllMarkets(ctx context.Context, in dto.ListAllMarketsInput) (*dto.ListAllMarketsOutput, error)
	UpdateMarketDetails(ctx context.Context, in dto.UpdateMarketDetailsInput) (*dto.UpdateMarketDetailsOutput, error)
	UpdateMarketRoles(ctx context.Context, in dto.UpdateMarketRolesInput) (*dto.UpdateMarketRolesOutput, error)
}
