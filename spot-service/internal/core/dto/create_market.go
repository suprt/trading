package dto

import "github.com/suprt/trading/spot-service/internal/core/domain"

type CreateMarketInput struct {
	CallerID      domain.UserID
	Name          string
	RequiredRoles []domain.Role
	Enabled       bool
}

type CreateMarketOutput struct {
	Market *domain.Market
}
