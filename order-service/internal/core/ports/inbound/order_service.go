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
	AdminGetOrder(ctx context.Context, in dto.AdminGetOrderInput) (*dto.AdminGetOrderOutput, error)
	AdminGetOrderStatus(ctx context.Context, in dto.AdminGetOrderStatusInput) (*dto.AdminGetOrderStatusOutput, error)
	CancelOrder(ctx context.Context, in dto.CancelOrderInput) (*dto.CancelOrderOutput, error)
	AdminCancelOrder(ctx context.Context, in dto.AdminCancelOrderInput) (*dto.AdminCancelOrderOutput, error)
	ListOrders(ctx context.Context, in dto.ListOrdersInput) (*dto.ListOrdersOutput, error)
	AdminListOrders(ctx context.Context, in dto.AdminListOrdersInput) (*dto.AdminListOrdersOutput, error)
}
