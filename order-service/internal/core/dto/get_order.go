package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type GetOrderInput struct {
	OrderID domain.OrderID
}

type GetOrderOutput struct {
	Order *domain.Order
}
