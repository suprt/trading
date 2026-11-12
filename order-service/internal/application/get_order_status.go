package application

import (
	"context"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type GetOrderStatusUseCase struct {
	orders outbound.OrderRepository
}

func NewGetOrderStatusUseCase(orders outbound.OrderRepository) *GetOrderStatusUseCase {
	return &GetOrderStatusUseCase{orders: orders}
}

func (uc *GetOrderStatusUseCase) GetOrderStatus(ctx context.Context, in dto.GetOrderStatusInput) (*dto.GetOrderStatusOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	order, err := uc.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	if order.UserID != callerID {
		return nil, er.ErrForbidden
	}
	return &dto.GetOrderStatusOutput{
		OrderID:   order.ID,
		Status:    order.Status,
		UpdatedAt: order.UpdatedAt,
	}, nil
}
