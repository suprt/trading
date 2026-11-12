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

func setupAdminGetOrderStatusTest() (*mockOrderRepo, *mockUserClient, *application.AdminGetOrderStatusUseCase) {
	repo := newMockOrderRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	uc := application.NewAdminGetOrderStatusUseCase(repo, users)
	return repo, users, uc
}

func TestAdminGetOrderStatus_Success(t *testing.T) {
	repo, _, uc := setupAdminGetOrderStatusTest()
	repo.orders["order-1"] = &domain.Order{
		ID: "order-1", UserID: "other-user", Status: domain.OrderStatusCancelled,
	}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	out, err := uc.AdminGetOrderStatus(ctx, dto.AdminGetOrderStatusInput{OrderID: "order-1"})
	require.NoError(t, err)
	require.Equal(t, domain.OrderStatusCancelled, out.Status)
}

func TestAdminGetOrderStatus_Forbidden(t *testing.T) {
	repo, users, uc := setupAdminGetOrderStatusTest()
	users.roles = []domain.Role{domain.RoleUser}
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "other-user"}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.AdminGetOrderStatus(ctx, dto.AdminGetOrderStatusInput{OrderID: "order-1"})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestAdminGetOrderStatus_NotFound(t *testing.T) {
	_, _, uc := setupAdminGetOrderStatusTest()
	ctx := auth.WithUserID(context.Background(), "admin-1")

	_, err := uc.AdminGetOrderStatus(ctx, dto.AdminGetOrderStatusInput{OrderID: "missing"})
	require.ErrorIs(t, err, er.ErrOrderNotFound)
}
