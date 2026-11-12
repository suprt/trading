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

func setupListOrdersTest() (*mockOrderRepo, *application.ListOrdersUseCase) {
	repo := newMockOrderRepo()
	uc := application.NewListOrdersUseCase(repo)
	return repo, uc
}

func TestListOrders_Success(t *testing.T) {
	repo, uc := setupListOrdersTest()
	repo.orders["order-1"] = &domain.Order{ID: "order-1", UserID: "user-1"}
	repo.orders["order-2"] = &domain.Order{ID: "order-2", UserID: "user-1"}
	repo.orders["order-3"] = &domain.Order{ID: "order-3", UserID: "other-user"} // чужой
	ctx := auth.WithUserID(context.Background(), "user-1")

	out, err := uc.ListOrders(ctx, dto.ListOrdersInput{PageSize: 10})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"order-1", "order-2"}, orderNames(out.Orders))
}

func TestListOrders_InvalidPageToken(t *testing.T) {
	_, uc := setupListOrdersTest()
	ctx := auth.WithUserID(context.Background(), "user-1")

	_, err := uc.ListOrders(ctx, dto.ListOrdersInput{PageSize: 10, PageToken: "not-base64!!!"})
	require.ErrorIs(t, err, er.ErrInvalidPageToken)
}

func TestListOrders_NoAuth(t *testing.T) {
	_, uc := setupListOrdersTest()

	_, err := uc.ListOrders(context.Background(), dto.ListOrdersInput{PageSize: 10})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
