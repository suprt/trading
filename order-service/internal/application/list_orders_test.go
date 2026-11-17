package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/order-service/internal/application"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/dto"
	er "github.com/suprt/trading/order-service/internal/core/errors"
)

func setupListOrdersTest() (*mockOrderRepo, *application.ListOrdersUseCase) {
	repo := newMockOrderRepo()
	uc := application.NewListOrdersUseCase(repo)
	return repo, uc
}
func TestListOrders_User(t *testing.T) {
	repo, uc := setupListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "other-user"}
	ctx := ctxWithUser("user-1")

	out, err := uc.ListOrders(ctx, dto.ListOrdersInput{PageSize: 10})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1"}, orderNames(out.Orders))
}

func TestListOrders_AdminAll(t *testing.T) {
	repo, uc := setupListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "user-2"}
	ctx := ctxWithAdmin("admin-1")

	out, err := uc.ListOrders(ctx, dto.ListOrdersInput{PageSize: 10})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1", "order-2"}, orderNames(out.Orders))
}

func TestListOrders_AdminFilterByUser(t *testing.T) {
	repo, uc := setupListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "user-2"}
	ctx := ctxWithAdmin("admin-1")

	uid := domain.UserID("user-1")
	out, err := uc.ListOrders(ctx, dto.ListOrdersInput{UserID: &uid, PageSize: 10})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1"}, orderNames(out.Orders))
}

func TestListOrders_UserForbiddenOtherUserID(t *testing.T) {
	_, uc := setupListOrdersTest()
	ctx := ctxWithUser("user-1")

	other := domain.UserID("other-user")
	_, err := uc.ListOrders(ctx, dto.ListOrdersInput{UserID: &other, PageSize: 10})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestListOrders_InvalidPageToken(t *testing.T) {
	_, uc := setupListOrdersTest()
	ctx := ctxWithUser("user-1")
	_, err := uc.ListOrders(ctx, dto.ListOrdersInput{PageSize: 10, PageToken: "not-base64!!!"})
	require.ErrorIs(t, err, er.ErrInvalidPageToken)
}

func TestListOrders_NoAuth(t *testing.T) {
	_, uc := setupListOrdersTest()
	_, err := uc.ListOrders(context.Background(), dto.ListOrdersInput{PageSize: 10})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
