package dto

import (
	"time"

	"github.com/suprt/trading/order-service/internal/core/domain"
)

type AdminGetOrderStatusInput struct {
	OrderID domain.OrderID
}
type AdminGetOrderStatusOutput struct {
	OrderID   domain.OrderID
	Status    domain.OrderStatus
	UpdatedAt time.Time
}
