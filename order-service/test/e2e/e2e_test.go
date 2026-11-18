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
	otherID  = "00000000-0000-0000-0000-000000000004"
)

func defaultMarkets() map[domain.MarketID]*domain.Market {
	return map[domain.MarketID]*domain.Market{
		marketID: {
			ID:            marketID,
			Name:          "BTC/USDT",
			RequiredRoles: nil,
		},
	}
}

func TestE2E_CreateOrder_Success(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	price := int64(60000)
	resp, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
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
		RequiredRoles: []domain.Role{domain.RoleAdmin},
	}
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, markets, 100)
	defer env.cleanup()

	resp, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)
	require.Equal(t, orderv1.OrderStatus_ORDER_STATUS_REJECTED, resp.Order.Status)
}

func TestE2E_CreateOrder_MarketNotFound(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	_, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
		MarketId: "00000000-0000-0000-0000-000000000099",
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestE2E_CreateOrder_Unauthenticated(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
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
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	ctx := authCtx(t, userID, []string{"USER"})
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

	_, err = env.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestE2E_CancelOrder_Forbidden(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	_, err = env.client.CancelOrder(authCtx(t, otherID, []string{"USER"}), &orderv1.CancelOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_GetOrder_Forbidden(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	_, err = env.client.GetOrder(authCtx(t, otherID, []string{"USER"}), &orderv1.GetOrderRequest{
		OrderId: created.Order.Id,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_ListOrders(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	ctx := authCtx(t, userID, []string{"USER"})
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
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 100)
	defer env.cleanup()

	_, err := env.client.ListOrders(authCtx(t, userID, []string{"USER"}), &orderv1.ListOrdersRequest{
		PageSize:  10,
		PageToken: "not-base64!!!",
	})
	require.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestE2E_ListOrders_AdminFilterByUser(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser, domain.RoleAdmin}, defaultMarkets(), 100)
	defer env.cleanup()

	for i := 0; i < 2; i++ {
		_, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
			MarketId: string(marketID),
			Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
			Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
			Quantity: 1,
		})
		require.NoError(t, err)
	}

	uid := userID
	out, err := env.client.ListOrders(authCtx(t, adminID, []string{"ADMIN"}), &orderv1.ListOrdersRequest{
		UserId:   &uid,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.Len(t, out.Orders, 2)
	for _, o := range out.Orders {
		require.Equal(t, userID, o.UserId)
	}
}

func TestE2E_GetOrder_AdminCanSeeOthers(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser, domain.RoleAdmin}, defaultMarkets(), 100)
	defer env.cleanup()

	created, err := env.client.CreateOrder(authCtx(t, userID, []string{"USER"}), &orderv1.CreateOrderRequest{
		MarketId: string(marketID),
		Side:     orderv1.OrderSide_ORDER_SIDE_BUY,
		Type:     orderv1.OrderType_ORDER_TYPE_MARKET,
		Quantity: 1,
	})
	require.NoError(t, err)

	got, err := env.client.GetOrder(authCtx(t, adminID, []string{"ADMIN"}), &orderv1.GetOrderRequest{
		OrderId: created.Order.Id,
	})
	require.NoError(t, err)
	require.Equal(t, created.Order.Id, got.Order.Id)
}

func TestE2E_RateLimit_Basic(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleUser}, defaultMarkets(), 2) // perIPRPM=2
	defer env.cleanup()

	ctx := authCtx(t, userID, []string{"USER"})

	for i := 0; i < 2; i++ {
		_, err := env.client.GetOrder(ctx, &orderv1.GetOrderRequest{
			OrderId: "00000000-0000-0000-0000-000000000099",
		})
		require.NotEqual(t, codes.ResourceExhausted, status.Code(err))
	}

	_, err := env.client.GetOrder(ctx, &orderv1.GetOrderRequest{
		OrderId: "00000000-0000-0000-0000-000000000099",
	})
	require.Error(t, err)
	require.Equal(t, codes.ResourceExhausted, status.Code(err))
	require.Equal(t, "too many requests", status.Convert(err).Message())
}

func TestE2E_RateLimit_AdminSkip(t *testing.T) {
	env := setupTestServer(t, []domain.Role{domain.RoleAdmin}, defaultMarkets(), 2) // perIPRPM=2, но admin skip
	defer env.cleanup()

	ctx := authCtx(t, adminID, []string{"ADMIN"})

	for i := 0; i < 10; i++ {
		_, err := env.client.GetOrder(ctx, &orderv1.GetOrderRequest{
			OrderId: "00000000-0000-0000-0000-000000000099",
		})
		require.NotEqual(t, codes.ResourceExhausted, status.Code(err))
	}
}
