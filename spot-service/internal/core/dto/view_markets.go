package dto

import (
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type ViewMarketsInput struct{}

type ViewMarketsOutput struct {
	Markets []*domain.Market
}
