package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type ListOrdersInput struct {
	PageSize  int32
	PageToken string
}

type ListOrdersOutput struct {
	Orders        []*domain.Order
	NextPageToken string
}
