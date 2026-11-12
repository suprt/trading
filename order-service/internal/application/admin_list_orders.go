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

type AdminListOrdersUseCase struct {
	orders outbound.OrderRepository
	users  outbound.UserClient
}

func NewAdminListOrdersUseCase(orders outbound.OrderRepository, users outbound.UserClient) *AdminListOrdersUseCase {
	return &AdminListOrdersUseCase{orders: orders, users: users}
}

func (uc *AdminListOrdersUseCase) AdminListOrders(ctx context.Context, in dto.AdminListOrdersInput) (*dto.AdminListOrdersOutput, error) {
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

	limit := int(in.PageSize)
	if limit <= 0 {
		limit = 20
	}

	var afterCreatedAt time.Time
	var afterID domain.OrderID
	if in.PageToken != "" {
		afterCreatedAt, afterID, err = decode(in.PageToken)
		if err != nil {
			return nil, er.ErrInvalidPageToken
		}
	}

	orders, err := uc.orders.ListAll(ctx, in.UserID, limit+1, afterCreatedAt, afterID)
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if len(orders) > limit {
		orders = orders[:limit]
		last := orders[len(orders)-1]
		nextPageToken = encode(last.CreatedAt, last.ID)
	}

	return &dto.AdminListOrdersOutput{
		Orders:        orders,
		NextPageToken: nextPageToken,
	}, nil
}
