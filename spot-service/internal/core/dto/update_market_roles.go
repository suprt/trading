package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type UpdateMarketRolesInput struct {
	CallerID      domain.UserID
	ID            domain.MarketID
	RequiredRoles []domain.Role
}
type UpdateMarketRolesOutput struct {
	Market *domain.Market
}
