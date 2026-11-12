package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type ViewMarketsInput struct {
	UserRoles []domain.Role
}

type ViewMarketsOutput struct {
	Markets []*domain.Market
}
