package dto

import (
	"time"

	"github.com/suprt/trading/order-service/internal/core/domain"
)

type GetOrderStatusInput struct {
	OrderID domain.OrderID
}
type GetOrderStatusOutput struct {
	OrderID   domain.OrderID
	Status    domain.OrderStatus
	UpdatedAt time.Time
}
