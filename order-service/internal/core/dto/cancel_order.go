package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type CancelOrderInput struct {
	OrderID domain.OrderID
}

type CancelOrderOutput struct {
	Order *domain.Order
}
