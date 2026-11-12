package application

import (
	"context"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type GetOrderUseCase struct {
	orders outbound.OrderRepository
}

func NewGetOrderUseCase(orders outbound.OrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{orders: orders}
}

func (uc *GetOrderUseCase) GetOrder(ctx context.Context, in dto.GetOrderInput) (*dto.GetOrderOutput, error) {
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
	return &dto.GetOrderOutput{Order: order}, nil
}
