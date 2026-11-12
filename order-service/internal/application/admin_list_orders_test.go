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

func setupAdminListOrdersTest() (*mockOrderRepo, *mockUserClient, *application.AdminListOrdersUseCase) {
	repo := newMockOrderRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	uc := application.NewAdminListOrdersUseCase(repo, users)
	return repo, users, uc
}

func TestAdminListOrders_All(t *testing.T) {
	repo, _, uc := setupAdminListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "user-2"}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	out, err := uc.AdminListOrders(ctx, dto.AdminListOrdersInput{PageSize: 10})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1", "order-2"}, orderNames(out.Orders))
}

func TestAdminListOrders_FilterByUser(t *testing.T) {
	repo, _, uc := setupAdminListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "user-2"}
	ctx := auth.WithUserID(context.Background(), "admin-1")

	uid := domain.UserID("user-1")
	out, err := uc.AdminListOrders(ctx, dto.AdminListOrdersInput{
		UserID: &uid, PageSize: 10,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1"}, orderNames(out.Orders))
}

func TestAdminListOrders_Forbidden(t *testing.T) {
	repo, users, uc := setupAdminListOrdersTest()
	users.roles = []domain.Role{domain.RoleUser}
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.AdminListOrders(ctx, dto.AdminListOrdersInput{PageSize: 10})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestAdminListOrders_InvalidPageToken(t *testing.T) {
	_, _, uc := setupAdminListOrdersTest()
	ctx := auth.WithUserID(context.Background(), "admin-1")

	_, err := uc.AdminListOrders(ctx, dto.AdminListOrdersInput{PageSize: 10, PageToken: "bad"})
	require.ErrorIs(t, err, er.ErrInvalidPageToken)
}
