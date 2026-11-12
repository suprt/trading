package grpc_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	grpchand "github.com/suprt/trading/order-service/internal/adapters/inbound/grpc"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	"github.com/suprt/trading/order-service/internal/core/ports/inbound/mocks"
	orderv1 "github.com/suprt/trading/proto/order/v1"
)

func newTestOrder() *domain.Order {
	return &domain.Order{
		ID:       "00000000-0000-0000-0000-000000000001",
		UserID:   "00000000-0000-0000-0000-000000000002",
		MarketID: "00000000-0000-0000-0000-000000000003",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeLimit,
		Price:    60000,
		Quantity: 2,
		Status:   domain.OrderStatusCreated,
	}
}

func TestHandler_CreateOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	price := int64(60000)
	svc.EXPECT().
		CreateOrder(gomock.Any(), dto.CreateOrderInput{
			MarketID: "00000000-0000-0000-0000-000000000003",
			Side:     domain.OrderSideBuy,
			Type:     domain.OrderTypeLimit,
			Price:    &price,
			Quantity: 2,
		}).
		Return(&dto.CreateOrderOutput{Order: newTestOrder()}, nil)

	h := grpchand.New(svc)

	resp, err := h.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		MarketId: "00000000-0000-0000-0000-000000000003",
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_LIMIT,
		Price:    &price,
		Quantity: 2,
	})
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", resp.Order.Id)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CREATED, resp.Order.Status)
}

func TestHandler_CreateOrder_InvalidSide(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	h := grpchand.New(svc)

	_, err := h.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		MarketId: "00000000-0000-0000-0000-000000000003",
		Side:     orderv1.OrderSide_ORDER_SIDE_UNSPECIFIED,
		Type:     orderv1.OrderType_ORDER_TYPE_LIMIT,
		Quantity: 2,
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandler_CreateOrder_InvalidPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().
		CreateOrder(gomock.Any(), gomock.Any()).
		Return(nil, er.ErrInvalidPrice)

	h := grpchand.New(svc)

	_, err := h.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		MarketId: "00000000-0000-0000-0000-000000000003",
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_LIMIT,
		Quantity: 2,
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestHandler_GetOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().
		GetOrder(gomock.Any(), dto.GetOrderInput{
			OrderID: "00000000-0000-0000-0000-000000000001",
		}).
		Return(&dto.GetOrderOutput{Order: newTestOrder()}, nil)

	h := grpchand.New(svc)
	resp, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.NoError(t, err)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", resp.Order.Id)
}

func TestHandler_GetOrder_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Return(nil, er.ErrOrderNotFound)

	h := grpchand.New(svc)
	_, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000099",
	})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestHandler_GetOrder_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().GetOrder(gomock.Any(), gomock.Any()).Return(nil, er.ErrForbidden)

	h := grpchand.New(svc)
	_, err := h.GetOrder(context.Background(), &orderv1.GetOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestHandler_GetOrderStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().
		GetOrderStatus(gomock.Any(), dto.GetOrderStatusInput{
			OrderID: "00000000-0000-0000-0000-000000000001",
		}).
		Return(&dto.GetOrderStatusOutput{
			OrderID: "00000000-0000-0000-0000-000000000001",
			Status:  domain.OrderStatusCancelled,
		}, nil)

	h := grpchand.New(svc)
	resp, err := h.GetOrderStatus(context.Background(), &orderv1.GetOrderStatusRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CANCELLED, resp.Status)
	require.Equal(t, "00000000-0000-0000-0000-000000000001", resp.OrderId)
}

func TestHandler_CancelOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	cancelled := newTestOrder()
	cancelled.Status = domain.OrderStatusCancelled
	svc.EXPECT().
		CancelOrder(gomock.Any(), dto.CancelOrderInput{
			OrderID: "00000000-0000-0000-0000-000000000001",
		}).
		Return(&dto.CancelOrderOutput{Order: cancelled}, nil)

	h := grpchand.New(svc)
	resp, err := h.CancelOrder(context.Background(), &orderv1.CancelOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CANCELLED, resp.Order.Status)
}

func TestHandler_CancelOrder_NotCancellable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().CancelOrder(gomock.Any(), gomock.Any()).Return(nil, er.ErrOrderNotCancellable)

	h := grpchand.New(svc)
	_, err := h.CancelOrder(context.Background(), &orderv1.CancelOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestHandler_AdminCancelOrder_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	cancelled := newTestOrder()
	cancelled.Status = domain.OrderStatusCancelled
	svc.EXPECT().
		AdminCancelOrder(gomock.Any(), dto.AdminCancelOrderInput{
			OrderID: "00000000-0000-0000-0000-000000000001",
		}).
		Return(&dto.AdminCancelOrderOutput{Order: cancelled}, nil)

	h := grpchand.New(svc)
	resp, err := h.AdminCancelOrder(context.Background(), &orderv1.AdminCancelOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CANCELLED, resp.Order.Status)
}

func TestHandler_AdminCancelOrder_Forbidden(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().AdminCancelOrder(gomock.Any(), gomock.Any()).Return(nil, er.ErrForbidden)

	h := grpchand.New(svc)
	_, err := h.AdminCancelOrder(context.Background(), &orderv1.AdminCancelOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000001",
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestHandler_ListOrders_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().
		ListOrders(gomock.Any(), dto.ListOrdersInput{PageSize: 10}).
		Return(&dto.ListOrdersOutput{
			Orders:        []*domain.Order{newTestOrder()},
			NextPageToken: "next",
		}, nil)

	h := grpchand.New(svc)
	resp, err := h.ListOrders(context.Background(), &orderv1.ListOrdersRequest{PageSize: 10})
	require.NoError(t, err)
	require.Len(t, resp.Orders, 1)
	require.Equal(t, "next", resp.NextPageToken)
}

func TestHandler_AdminListOrders_FilterByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	uid := domain.UserID("00000000-0000-0000-0000-000000000002")
	svc := mocks.NewMockOrderService(ctrl)
	svc.EXPECT().
		AdminListOrders(gomock.Any(), dto.AdminListOrdersInput{
			UserID:   &uid,
			PageSize: 10,
		}).
		Return(&dto.AdminListOrdersOutput{Orders: []*domain.Order{newTestOrder()}}, nil)

	h := grpchand.New(svc)
	userIdStr := "00000000-0000-0000-0000-000000000002"
	resp, err := h.AdminListOrders(context.Background(), &orderv1.AdminListOrdersRequest{
		UserId:   &userIdStr,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, resp.Orders, 1)
}
