package dto

import "github.com/suprt/trading/spot-service/internal/core/domain"

type DeleteMarketInput struct {
	CallerID domain.UserID
	ID       domain.MarketID
}
