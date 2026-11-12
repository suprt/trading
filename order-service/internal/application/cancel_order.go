package application

import (
	"context"
	"time"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type CancelOrderUseCase struct {
	orders outbound.OrderRepository
}

func NewCancelOrderUseCase(orders outbound.OrderRepository) *CancelOrderUseCase {
	return &CancelOrderUseCase{orders: orders}
}

func (uc *CancelOrderUseCase) CancelOrder(ctx context.Context, in dto.CancelOrderInput) (*dto.CancelOrderOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	order, updated, err := uc.orders.Cancel(ctx, in.OrderID, &callerID, time.Now())
	if err != nil {
		return nil, err
	}
	if !updated {
		order, err := uc.orders.GetByID(ctx, in.OrderID)
		if err != nil {
			return nil, err
		}
		if order.UserID != callerID {
			return nil, er.ErrForbidden
		}
		return nil, er.ErrOrderNotCancellable
	}
	return &dto.CancelOrderOutput{Order: order}, nil
}
