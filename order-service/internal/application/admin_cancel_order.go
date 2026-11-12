package application

import (
	"context"
	"slices"
	"time"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type AdminCancelOrderUseCase struct {
	orders outbound.OrderRepository
	users  outbound.UserClient
}

func NewAdminCancelOrderUseCase(orders outbound.OrderRepository, users outbound.UserClient) *AdminCancelOrderUseCase {
	return &AdminCancelOrderUseCase{orders: orders, users: users}
}

func (uc *AdminCancelOrderUseCase) AdminCancelOrder(ctx context.Context, in dto.AdminCancelOrderInput) (*dto.AdminCancelOrderOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	roles, _, err := uc.users.GetUserRoles(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if !slices.Contains(roles, domain.RoleAdmin) {
		return nil, er.ErrForbidden
	}

	order, updated, err := uc.orders.Cancel(ctx, in.OrderID, nil, time.Now())
	if err != nil {
		return nil, err
	}
	if !updated {

		if _, getErr := uc.orders.GetByID(ctx, in.OrderID); getErr != nil {
			return nil, getErr
		}
		return nil, er.ErrOrderNotCancellable
	}
	return &dto.AdminCancelOrderOutput{Order: order}, nil
}