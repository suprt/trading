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

type CreateOrderUseCase struct {
	orders outbound.OrderRepository
	users  outbound.UserClient
	spots  outbound.SpotClient
	ids    outbound.IDGenerator
}

func NewCreateOrderUseCase(
	orders outbound.OrderRepository,
	users outbound.UserClient,
	spots outbound.SpotClient,
	ids outbound.IDGenerator,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orders: orders,
		users:  users,
		spots:  spots,
		ids:    ids,
	}
}

func (uc *CreateOrderUseCase) CreateOrder(ctx context.Context, in dto.CreateOrderInput) (*dto.CreateOrderOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}

	if err := validatePrice(in.Type, in.Price); err != nil {
		return nil, err
	}

	market, err := uc.spots.GetMarket(ctx, in.MarketID)
	if err != nil {
		return nil, err
	}

	roles, _, err := uc.users.GetUserRoles(ctx, callerID)
	if err != nil {
		return nil, err
	}

	status := domain.OrderStatusCreated
	if !market.AvailableFor(roles) {
		status = domain.OrderStatusRejected
	}

	var price int64
	if in.Type == domain.OrderTypeLimit {
		price = *in.Price
	}
	now := time.Now()
	order := &domain.Order{
		ID:        domain.OrderID(uc.ids.New()),
		UserID:    callerID,
		MarketID:  in.MarketID,
		Side:      in.Side,
		Type:      in.Type,
		Price:     price,
		Quantity:  in.Quantity,
		Status:    status,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.orders.Create(ctx, order); err != nil {
		return nil, err
	}

	return &dto.CreateOrderOutput{Order: order}, nil
}

func validatePrice(t domain.OrderType, price *int64) error {
	switch t {
	case domain.OrderTypeLimit:
		if price == nil || *price <= 0 {
			return er.ErrInvalidPrice
		}
	case domain.OrderTypeMarket:

	default:
		return er.ErrInvalidOrderType
	}
	return nil
}