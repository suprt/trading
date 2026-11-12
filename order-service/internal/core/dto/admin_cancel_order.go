package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type AdminCancelOrderInput struct {
	OrderID domain.OrderID
}

type AdminCancelOrderOutput struct {
	Order *domain.Order
}
