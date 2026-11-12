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

func setupCancelOrderTest() (*mockOrderRepo, *application.CancelOrderUseCase) {
	repo := newMockOrderRepo()
	uc := application.NewCancelOrderUseCase(repo)
	return repo, uc
}

func TestCancelOrder_Success(t *testing.T) {
	repo, uc := setupCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCancelled, out.Order.Status)

	saved := repo.orders["order-1"]
	require.Equal(t, domain.OrderStatusCancelled, saved.Status)
}

func TestCancelOrder_NotFound(t *testing.T) {
	_, uc := setupCancelOrderTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestCancelOrder_Forbidden(t *testing.T) {
	repo, uc := setupCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "other-user",
		Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)

	// статус не изменился
	require.Equal(t, domain.OrderStatusCreated, repo.orders["order-1"].Status)
}

func TestCancelOrder_NotCancellable_AlreadyCancelled(t *testing.T) {
	repo, uc := setupCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.OrderStatusCancelled,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrOrderNotCancellable)
}

func TestCancelOrder_NotCancellable_Filled(t *testing.T) {
	repo, uc := setupCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.OrderStatusFilled,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrOrderNotCancellable)
}

func TestCancelOrder_NotCancellable_Rejected(t *testing.T) {
	repo, uc := setupCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "user-1",
		Status: domain.OrderStatusRejected,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.CancelOrder(ctx, dto.CancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrOrderNotCancellable)
}

func TestCancelOrder_NoAuth(t *testing.T) {
	_, uc := setupCancelOrderTest()

	_, err := uc.CancelOrder(context.Background(), dto.CancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
