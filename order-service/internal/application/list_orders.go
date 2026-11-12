package application

import (
	"context"
	"time"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
)

type ListOrdersUseCase struct {
	orders outbound.OrderRepository
}

func NewListOrdersUseCase(orders outbound.OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{orders: orders}
}

func (uc *ListOrdersUseCase) ListOrders(ctx context.Context, in dto.ListOrdersInput) (*dto.ListOrdersOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	limit := int(in.PageSize)
	if limit <= 0 {
		limit = 20
	}

	var afterCreatedAt time.Time
	var afterID domain.OrderID
	if in.PageToken != "" {
		var err error
		afterCreatedAt, afterID, err = decode(in.PageToken)
		if err != nil {
			return nil, er.ErrInvalidPageToken
		}
	}

	orders, err := uc.orders.ListByUser(ctx, callerID, limit+1, afterCreatedAt, afterID)
	if err != nil {
		return nil, err
	}

	var nextPageToken string
	if len(orders) > limit {
		orders = orders[:limit]
		last := orders[len(orders)-1]
		nextPageToken = encode(last.CreatedAt, last.ID)
	}

	return &dto.ListOrdersOutput{
		Orders:        orders,
		NextPageToken: nextPageToken,
	}, nil
}
