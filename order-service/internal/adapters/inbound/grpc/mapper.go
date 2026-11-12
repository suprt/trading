package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	orderv1 "github.com/suprt/trading/proto/order/v1"
)

func orderToProto(o *domain.Order) *orderv1.Order {
	return &orderv1.Order{
		Id:        string(o.ID),
		UserId:    string(o.UserID),
		MarketId:  string(o.MarketID),
		Side:      sideToProto(o.Side),
		Type:      typeToProto(o.Type),
		Price:     o.Price,
		Quantity:  o.Quantity,
		Status:    statusToProto(o.Status),
		CreatedAt: timestamppb.New(o.CreatedAt),
		UpdatedAt: timestamppb.New(o.UpdatedAt),
	}
}

func sideToProto(s domain.OrderSide) orderv1.OrderSide {
	switch s {
	case domain.OrderSideBuy:
		return orderv1.OrderSide_ORDER_SIDE_BUY
	case domain.OrderSideSell:
		return orderv1.OrderSide_ORDER_SIDE_SELL
	default:
		return orderv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func typeToProto(t domain.OrderType) orderv1.OrderType {
	switch t {
	case domain.OrderTypeLimit:
		return orderv1.OrderType_ORDER_TYPE_LIMIT
	case domain.OrderTypeMarket:
		return orderv1.OrderType_ORDER_TYPE_MARKET
	default:
		return orderv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

func statusToProto(s domain.OrderStatus) orderv1.OrderStatus {
	switch s {
	case domain.OrderStatusCreated:
		return orderv1.OrderStatus_ORDER_STATUS_CREATED
	case domain.OrderStatusFilled:
		return orderv1.OrderStatus_ORDER_STATUS_FILLED
	case domain.OrderStatusCancelled:
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	case domain.OrderStatusRejected:
		return orderv1.OrderStatus_ORDER_STATUS_REJECTED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func sideFromProto(s orderv1.OrderSide) (domain.OrderSide, error) {
	switch s {
	case orderv1.OrderSide_ORDER_SIDE_BUY:
		return domain.OrderSideBuy, nil
	case orderv1.OrderSide_ORDER_SIDE_SELL:
		return domain.OrderSideSell, nil
	default:
		return "", er.ErrInvalidOrderSide
	}
}

func typeFromProto(t orderv1.OrderType) (domain.OrderType, error) {
	switch t {
	case orderv1.OrderType_ORDER_TYPE_LIMIT:
		return domain.OrderTypeLimit, nil
	case orderv1.OrderType_ORDER_TYPE_MARKET:
		return domain.OrderTypeMarket, nil
	default:
		return "", er.ErrInvalidOrderType
	}
}
