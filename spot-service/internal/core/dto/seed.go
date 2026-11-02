package dto

import "github.com/suprt/trading/spot-service/internal/core/domain"

type SeedMarketInput struct {
	Name          string
	RequiredRoles []domain.Role
	Enabled       bool
}
