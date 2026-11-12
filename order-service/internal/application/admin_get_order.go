package application

import (
	"context"
	"slices"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type AdminGetOrderUseCase struct {
	orders outbound.OrderRepository
	users  outbound.UserClient
}

func NewAdminGetOrderUseCase(orders outbound.OrderRepository, users outbound.UserClient) *AdminGetOrderUseCase {
	return &AdminGetOrderUseCase{orders: orders, users: users}
}

func (uc *AdminGetOrderUseCase) AdminGetOrder(ctx context.Context, in dto.AdminGetOrderInput) (*dto.AdminGetOrderOutput, error) {
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

	order, err := uc.orders.GetByID(ctx, in.OrderID)
	if err != nil {
		return nil, err
	}
	return &dto.AdminGetOrderOutput{Order: order}, nil
}
