package dto

import "github.com/suprt/trading/spot-service/internal/core/domain"

type UpdateMarketDetailsInput struct {
	CallerID domain.UserID
	ID       domain.MarketID
	Name     *string
	Enabled  *bool
}
type UpdateMarketDetailsOutput struct {
	Market *domain.Market
}
