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

func setupGetOrderStatusTest() (*mockOrderRepo, *application.GetOrderStatusUseCase) {
	repo := newMockOrderRepo()
	uc := application.NewGetOrderStatusUseCase(repo)
	return repo, uc
}

func TestGetOrderStatus_Success(t *testing.T) {
	repo, uc := setupGetOrderStatusTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "user-1", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.GetOrderStatus(ctx, dto.GetOrderStatusInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderID("order-1"), out.OrderID)
	require.Equal(t, domain.OrderStatusCreated, out.Status)
}

func TestGetOrderStatus_NotFound(t *testing.T) {
	_, uc := setupGetOrderStatusTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.GetOrderStatus(ctx, dto.GetOrderStatusInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestGetOrderStatus_Forbidden(t *testing.T) {
	repo, uc := setupGetOrderStatusTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "other-user", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.GetOrderStatus(ctx, dto.GetOrderStatusInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestGetOrderStatus_NoAuth(t *testing.T) {
	_, uc := setupGetOrderStatusTest()

	_, err := uc.GetOrderStatus(context.Background(), dto.GetOrderStatusInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
