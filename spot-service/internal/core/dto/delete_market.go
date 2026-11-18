package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type DeleteMarketInput struct {
	ID domain.MarketID
}
