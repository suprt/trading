package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type CreateOrderInput struct {
	MarketID domain.MarketID
	Side     domain.OrderSide
	Type     domain.OrderType
	Price    *int64 // nil для MARKET
	Quantity int64
}

type CreateOrderOutput struct {
	Order *domain.Order
}
