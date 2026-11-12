package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type ListAllMarketsInput struct {
	CallerID        domain.UserID
	IncludeDeleted  bool
	IncludeDisabled bool
}

type ListAllMarketsOutput struct {
	Markets []*domain.Market
}
