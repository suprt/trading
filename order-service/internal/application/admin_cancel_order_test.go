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

func setupAdminCancelOrderTest() (*mockOrderRepo, *mockUserClient, *application.AdminCancelOrderUseCase) {
	repo := newMockOrderRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	uc := application.NewAdminCancelOrderUseCase(repo, users)
	return repo, users, uc
}

func TestAdminCancelOrder_Success(t *testing.T) {
	repo, _, uc := setupAdminCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "other-user", // чужой заказ
		Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	out, err := uc.AdminCancelOrder(ctx, dto.AdminCancelOrderInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCancelled, out.Order.Status)
	require.Equal(t, domain.OrderStatusCancelled, repo.orders["order-1"].Status)
}

func TestAdminCancelOrder_Forbidden(t *testing.T) {
	repo, users, uc := setupAdminCancelOrderTest()
	users.roles = []domain.Role{domain.RoleUser} // не админ
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "other-user",
		Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.AdminCancelOrder(ctx, dto.AdminCancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)
	require.Equal(t, domain.OrderStatusCreated, repo.orders["order-1"].Status)
}

func TestAdminCancelOrder_NotFound(t *testing.T) {
	_, _, uc := setupAdminCancelOrderTest()
	ctx := auth.WithUserID(context.Background(), "admin-1")

	_, err := uc.AdminCancelOrder(ctx, dto.AdminCancelOrderInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestAdminCancelOrder_NotCancellable(t *testing.T) {
	repo, _, uc := setupAdminCancelOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID:     "order-1",
		UserID: "other-user",
		Status: domain.OrderStatusFilled,
	}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	_, err := uc.AdminCancelOrder(ctx, dto.AdminCancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrOrderNotCancellable)
}

func TestAdminCancelOrder_NoAuth(t *testing.T) {
	_, _, uc := setupAdminCancelOrderTest()

	_, err := uc.AdminCancelOrder(context.Background(), dto.AdminCancelOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
