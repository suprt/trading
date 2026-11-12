package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/order-service/internal/application"
	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
)

func setupCreateOrderTest() (*mockOrderRepo, *mockUserClient, *mockSpotClient, *application.CreateOrderUseCase) {
	repo := newMockOrderRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleUser}}
	spots := &mockSpotClient{market: &domain.Market{
		ID:            "market-1",
		Name:          "BTC/USDT",
		RequiredRoles: nil, // доступен всем
	}}
	ids := &mockIDGen{id: "order-1"}

	uc := application.NewCreateOrderUseCase(repo, users, spots, ids)
	return repo, users, spots, uc
}

func TestCreateOrder_LimitSuccess(t *testing.T) {
	repo, _, _, uc := setupCreateOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeLimit,
		Price:    ptr(int64(60000)),
		Quantity: 2,
	})
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCreated, out.Order.Status)
	require.Equal(t, int64(60000), out.Order.Price)
	require.Equal(t, int64(2), out.Order.Quantity)

	_, ok := repo.orders["order-1"]
	require.True(t, ok)
}

func TestCreateOrder_MarketPriceZero(t *testing.T) {
	_, _, _, uc := setupCreateOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeMarket,
		Quantity: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), out.Order.Price) // MARKET → price 0
}

func TestCreateOrder_MarketIgnoresPrice(t *testing.T) {
	_, _, _, uc := setupCreateOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeMarket,
		Price:    ptr(int64(60000)), // передан, но должен игнорироваться
		Quantity: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), out.Order.Price)
}

func TestCreateOrder_LimitWithoutPrice(t *testing.T) {
	_, _, _, uc := setupCreateOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeLimit,
		Price:    nil,
		Quantity: 2,
	})
	require.ErrorIs(t, err, er.ErrInvalidPrice)
}

func TestCreateOrder_LimitZeroPrice(t *testing.T) {
	_, _, _, uc := setupCreateOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeLimit,
		Price:    ptr(int64(0)),
		Quantity: 2,
	})
	require.ErrorIs(t, err, er.ErrInvalidPrice)
}

func TestCreateOrder_MarketNotFound(t *testing.T) {
	repo, _, spots, uc := setupCreateOrderTest()
	spots.err = er.ErrMarketNotFound
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "missing",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeMarket,
		Quantity: 1,
	})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
	require.Empty(t, repo.orders) // заказ не создан
}

func TestCreateOrder_RejectedNoRole(t *testing.T) {
	repo, users, spots, uc := setupCreateOrderTest()
	users.roles = []domain.Role{domain.RoleUser}
	spots.market.RequiredRoles = []domain.Role{domain.RoleAdmin} // только админ
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.CreateOrder(ctx, dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeMarket,
		Quantity: 1,
	})
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusRejected, out.Order.Status)

	// REJECTED сохраняется в БД
	saved, ok := repo.orders["order-1"]
	require.True(t, ok)
	require.Equal(t, domain.OrderStatusRejected, saved.Status)
}

func TestCreateOrder_NoAuth(t *testing.T) {
	_, _, _, uc := setupCreateOrderTest()

	_, err := uc.CreateOrder(context.Background(), dto.CreateOrderInput{
		MarketID: "market-1",
		Side:     domain.OrderSideBuy,
		Type:     domain.OrderTypeMarket,
		Quantity: 1,
	})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
