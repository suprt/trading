package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type AdminGetOrderInput struct {
	OrderID domain.OrderID
}

type AdminGetOrderOutput struct {
	Order *domain.Order
}
