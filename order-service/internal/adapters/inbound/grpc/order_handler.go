package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	"github.com/suprt/trading/order-service/internal/core/ports/inbound"
	orderv1 "github.com/suprt/trading/proto/order/v1"
)

type Handler struct {
	orderv1.UnimplementedOrderServiceServer
	orderService inbound.OrderService
}

func New(orderService inbound.OrderService) *Handler {
	return &Handler{orderService: orderService}
}

func (h *Handler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	side, err := sideFromProto(req.GetSide())
	if err != nil {
		return nil, mapError(err)
	}
	orderType, err := typeFromProto(req.GetType())
	if err != nil {
		return nil, mapError(err)
	}

	out, err := h.orderService.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: domain.MarketID(req.GetMarketId()),
		Side:     side,
		Type:     orderType,
		Price:    req.Price,
		Quantity: req.GetQuantity(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CreateOrderResponse{Order: orderToProto(out.Order)}, nil
}

func (h *Handler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	out, err := h.orderService.GetOrder(ctx, dto.GetOrderInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.GetOrderResponse{Order: orderToProto(out.Order)}, nil
}

func (h *Handler) GetOrderStatus(ctx context.Context, req *orderv1.GetOrderStatusRequest) (*orderv1.GetOrderStatusResponse, error) {
	out, err := h.orderService.GetOrderStatus(ctx, dto.GetOrderStatusInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.GetOrderStatusResponse{
		OrderId:   string(out.OrderID),
		Status:    statusToProto(out.Status),
		UpdatedAt: timestamppb.New(out.UpdatedAt),
	}, nil
}

func (h *Handler) AdminGetOrder(ctx context.Context, req *orderv1.AdminGetOrderRequest) (*orderv1.AdminGetOrderResponse, error) {
	out, err := h.orderService.AdminGetOrder(ctx, dto.AdminGetOrderInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.AdminGetOrderResponse{Order: orderToProto(out.Order)}, nil
}

func (h *Handler) AdminGetOrderStatus(ctx context.Context, req *orderv1.AdminGetOrderStatusRequest) (*orderv1.AdminGetOrderStatusResponse, error) {
	out, err := h.orderService.AdminGetOrderStatus(ctx, dto.AdminGetOrderStatusInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.AdminGetOrderStatusResponse{
		OrderId:   string(out.OrderID),
		Status:    statusToProto(out.Status),
		UpdatedAt: timestamppb.New(out.UpdatedAt),
	}, nil
}

func (h *Handler) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	out, err := h.orderService.CancelOrder(ctx, dto.CancelOrderInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.CancelOrderResponse{Order: orderToProto(out.Order)}, nil
}

func (h *Handler) AdminCancelOrder(ctx context.Context, req *orderv1.AdminCancelOrderRequest) (*orderv1.AdminCancelOrderResponse, error) {
	out, err := h.orderService.AdminCancelOrder(ctx, dto.AdminCancelOrderInput{
		OrderID: domain.OrderID(req.GetOrderId()),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &orderv1.AdminCancelOrderResponse{Order: orderToProto(out.Order)}, nil
}

func (h *Handler) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	out, err := h.orderService.ListOrders(ctx, dto.ListOrdersInput{
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	orders := make([]*orderv1.Order, len(out.Orders))
	for i, o := range out.Orders {
		orders[i] = orderToProto(o)
	}
	return &orderv1.ListOrdersResponse{
		Orders:        orders,
		NextPageToken: out.NextPageToken,
	}, nil
}

func (h *Handler) AdminListOrders(ctx context.Context, req *orderv1.AdminListOrdersRequest) (*orderv1.AdminListOrdersResponse, error) {
	in := dto.AdminListOrdersInput{
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	}
	if req.UserId != nil {
		uid := domain.UserID(req.GetUserId())
		in.UserID = &uid
	}

	out, err := h.orderService.AdminListOrders(ctx, in)
	if err != nil {
		return nil, mapError(err)
	}

	orders := make([]*orderv1.Order, len(out.Orders))
	for i, o := range out.Orders {
		orders[i] = orderToProto(o)
	}
	return &orderv1.AdminListOrdersResponse{
		Orders:        orders,
		NextPageToken: out.NextPageToken,
	}, nil
}
