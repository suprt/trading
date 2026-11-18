package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type UpdateMarketRolesInput struct {
	ID            domain.MarketID
	RequiredRoles []domain.Role
}
type UpdateMarketRolesOutput struct {
	Market *domain.Market
}
