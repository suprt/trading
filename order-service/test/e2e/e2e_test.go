package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/suprt/trading/order-service/internal/core/domain"
	orderv1 "github.com/suprt/trading/proto/order/v1"
)

var (
	marketID = domain.MarketID("00000000-0000-0000-0000-000000000001")
	userID   = "00000000-0000-0000-0000-000000000002"
	adminID  = "00000000-0000-0000-0000-000000000003"
)

func defaultMarkets() map[domain.MarketID]*domain.Market {
	return map[domain.MarketID]*domain.Market{
		marketID: {
			ID:            marketID,
			Name:          "BTC/USDT",
			RequiredRoles: nil, // доступен всем
		},
	}
}

func TestE2E_CreateOrder_Success(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	price := int64(60000)
	resp, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_LIMIT,
		Price:    &price,
		Quantity: 2,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CREATED, resp.Order.Status)
	require.Equal(t, int64(60000), resp.Order.Price)
}

func TestE2E_CreateOrder_RejectedNoRole(t *testing.T) {
	markets := defaultMarkets()
	markets[marketID] = &domain.Market{
		ID:            marketID,
		Name:          "BTC/USDT",
		RequiredRoles: []domain.Role{domain.RoleAdmin}, // только админ
	}
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, markets)
	defer env.cleanup()

	resp, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_REJECTED, resp.Order.Status)
}

func TestE2E_CreateOrder_MarketNotFound(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	_, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: "00000000-0000-0000-0000-000000000099",
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestE2E_CreateOrder_Unauthenticated(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	_, err := env.client.CreateOrder(context.Background(), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestE2E_CreateThenCancel(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	ctx := authCtx(t, userID)
	created, err := env.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CREATED, created.Order.Status)

	cancelled, err := env.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CANCELLED, cancelled.Order.Status)

	// повторная отмена — FailedPrecondition
	_, err = env.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestE2E_CancelOrder_Forbidden(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	other := "00000000-0000-0000-0000-000000000004"
	_, err = env.client.CancelOrder(authCtx(t, other), &orderv1.CancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_GetOrder_Forbidden(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	other := "00000000-0000-0000-0000-000000000004"
	_, err = env.client.GetOrder(authCtx(t, other), &orderv1.GetOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_AdminCancelOrder(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser, domain.RoleAdmin}, defaultMarkets())
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	cancelled, err := env.client.AdminCancelOrder(authCtx(t, adminID), &orderv1.AdminCancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_CANCELLED, cancelled.Order.Status)
}

func TestE2E_ListOrders(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	ctx := authCtx(t, userID)
	for i := 0; i < 3; i++ {
		_, err := env.client.CreateOrder(ctx, &orderv1.CreateOrderRequest{
			MarketId: string(marketID),
			Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
			Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
			Quantity: 1,
		})
		require.NoError(t, err)
	}

	out, err := env.client.ListOrders(ctx, &orderv1.ListOrdersRequest{PageSize: 10})
	require.NoError(t, err)
	require.Len(t, out.Orders, 3)
}

func TestE2E_ListOrders_InvalidPageToken(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets())
	defer env.cleanup()

	_, err := env.client.ListOrders(authCtx(t, userID), &orderv1.ListOrdersRequest{
		PageSize:  10,
		PageToken: "not-base64!!!",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestE2E_AdminListOrders_FilterByUser(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser, domain.RoleAdmin}, defaultMarkets())
	defer env.cleanup()

	// user создаёт 2 заказа
	for i := 0; i < 2; i++ {
		_, err := env.client.CreateOrder(authCtx(t, userID), &orderv1.CreateOrderRequest{
			MarketId: string(marketID),
			Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
			Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
			Quantity: 1,
		})
		require.NoError(t, err)
	}

	// админ видит только заказы userID
	uid := userID
	out, err := env.client.AdminListOrders(authCtx(t, adminID), &orderv1.AdminListOrdersRequest{
		UserId:   &uid,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, out.Orders, 2)
	for _, o := range out.Orders {
		require.Equal(t, userID, o.UserId)
	}
}
