package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	spotv1 "github.com/suprt/trading/proto/spot/v1"
	userv1 "github.com/suprt/trading/proto/user/v1"
)

// authCtx — эмулирует наличие токена (для прохождения Auth интерсептора).
func authCtx(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer fake-token")
}

func TestE2E_ViewMarkets_Public(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	out, err := client.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{
		UserRoles: []userv1.Role{userv1.Role_ROLE_USER},
	})
	require.NoError(t, err)
	require.NotEmpty(t, out.Markets) // сидер создал BTC, ETH
	// SOL не должен быть виден юзеру
	for _, m := range out.Markets {
		require.NotEqual(t, "SOL/USDT", m.Name)
	}
}

func TestE2E_CreateMarket_Forbidden(t *testing.T) {
	// Мок UserClient возвращает роли USER (не ADMIN)
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	ctx := authCtx(context.Background())

	_, err := client.CreateMarket(ctx, &spotv1.CreateMarketRequest{
		Name:          "TEST/USDT",
		RequiredRoles: []userv1.Role{userv1.Role_ROLE_USER},
		Enabled:       true,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_CreateMarket_Success(t *testing.T) {
	// Мок UserClient возвращает ADMIN
	client, cleanup := setupTestServer(t, []string{"ADMIN"})
	defer cleanup()

	ctx := authCtx(context.Background())

	out, err := client.CreateMarket(ctx, &spotv1.CreateMarketRequest{
		Name:          "NEW/USDT",
		RequiredRoles: []userv1.Role{userv1.Role_ROLE_USER},
		Enabled:       true,
	})
	require.NoError(t, err)
	require.Equal(t, "NEW/USDT", out.Market.Name)
}

func TestE2E_DeleteMarket_Forbidden(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	// Сначала создаём (через админский мок в другом тесте — тут просто проверяем запрет на удаление)
	ctx := authCtx(context.Background())
	_, err := client.DeleteMarket(ctx, &spotv1.DeleteMarketRequest{MarketId: "00000000-0000-0000-0000-000000000000"})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
