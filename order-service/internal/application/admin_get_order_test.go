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

func setupAdminGetOrderTest() (*mockOrderRepo, *mockUserClient, *application.AdminGetOrderUseCase) {
	repo := newMockOrderRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	uc := application.NewAdminGetOrderUseCase(repo, users)
	return repo, users, uc
}

func TestAdminGetOrder_Success(t *testing.T) {
	repo, _, uc := setupAdminGetOrderTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "other-user", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	out, err := uc.AdminGetOrder(ctx, dto.AdminGetOrderInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderID("order-1"), out.Order.ID)
}

func TestAdminGetOrder_Forbidden(t *testing.T) {
	repo, users, uc := setupAdminGetOrderTest()
	users.roles = []domain.Role{domain.RoleUser}
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "other-user", Status: domain.OrderStatusCreated,
	}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.AdminGetOrder(ctx, dto.AdminGetOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestAdminGetOrder_NotFound(t *testing.T) {
	_, _, uc := setupAdminGetOrderTest()
	ctx := auth.WithUserID(context.Background(), "admin-1")

	_, err := uc.AdminGetOrder(ctx, dto.AdminGetOrderInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}

func TestAdminGetOrder_NoAuth(t *testing.T) {
	_, _, uc := setupAdminGetOrderTest()

	_, err := uc.AdminGetOrder(context.Background(), dto.AdminGetOrderInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
