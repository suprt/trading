//go:generate mockgen -source=order_service.go -destination=mocks/mock_order_service.go -package=mocks
package inbound

import (
	"context"

	"github.com/suprt/trading/order-service/internal/core/dto"
)

type OrderService interface {
	CreateOrder(ctx context.Context, in dto.CreateOrderInput) (*dto.CreateOrderOutput, error)
	GetOrder(ctx context.Context, in dto.GetOrderInput) (*dto.GetOrderOutput, error)
	GetOrderStatus(ctx context.Context, in dto.GetOrderStatusInput) (*dto.GetOrderStatusOutput, error)
	CancelOrder(ctx context.Context, in dto.CancelOrderInput) (*dto.CancelOrderOutput, error)
	ListOrders(ctx context.Context, in dto.ListOrdersInput) (*dto.ListOrdersOutput, error)
}
