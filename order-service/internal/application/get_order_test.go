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

func setupGetOrderTest() (*mockOrderRepo, *application.GetOrderUseCase) {
	repo := newMockOrderRepo()
	uc := application.NewGetOrderUseCase(repo)
	return repo, uc
}

func TestGetOrder_Success(t *testing.T) {
	repo, uc := setupGetOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "user-1", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.GetOrder(ctx, dto.GetOrderInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderID("order-1"), out.Order.ID)
}

func TestGetOrder_NotFound(t *testing.T) {
	_, uc := setupGetOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.GetOrder(ctx, dto.GetOrderInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestGetOrder_Forbidden(t *testing.T) {
	repo, uc := setupGetOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "other-user", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.GetOrder(ctx, dto.GetOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestGetOrder_NoAuth(t *testing.T) {
	_, uc := setupGetOrderTest()

	_, err := uc.GetOrder(context.Background(), dto.GetOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
