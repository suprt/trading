package dto

import "github.com/suprt/trading/spot-service/internal/core/domain"

type GetMarketInput struct {
	ID domain.MarketID
}

type GetMarketOutput struct {
	Market *domain.Market
}
