package e2e_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	userv1 "github.com/suprt/trading/proto/user/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	spotv1 "github.com/suprt/trading/proto/spot/v1"
)

func TestE2E_ViewMarkets_Authorized(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	out, err := client.ViewMarkets(authCtx(t, "user-1", []string{"USER"}), &spotv1.ViewMarketsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, out.Markets)
	// SOL/USDT требует TRADER — USER его не видит
	for _, m := range out.Markets {
		require.NotEqual(t, "SOL/USDT", m.Name)
	}
}

func TestE2E_ViewMarkets_Trader(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"TRADER"})
	defer cleanup()

	out, err := client.ViewMarkets(authCtx(t, "trader-1", []string{"TRADER"}), &spotv1.ViewMarketsRequest{})
	require.NoError(t, err)

	names := make([]string, len(out.Markets))
	for i, m := range out.Markets {
		names[i] = m.Name
	}
	require.Contains(t, names, "BTC/USDT")
	require.Contains(t, names, "SOL/USDT") // TRADER видит SOL
}

func TestE2E_ViewMarkets_Unauthenticated(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	// без токена — Unauthenticated (ViewMarkets больше не в whitelist)
	_, err := client.ViewMarkets(context.Background(), &spotv1.ViewMarketsRequest{})
	require.Error(t, err)
	assert.Equal(t, codes.Unauthenticated, status.Code(err))
}

func TestE2E_GetMarket_Public(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	// GetMarket в whitelist — без токена
	out, err := client.ViewMarkets(authCtx(t, "user-1", []string{"USER"}), &spotv1.ViewMarketsRequest{})
	require.NoError(t, err)
	require.NotEmpty(t, out.Markets)

	marketID := out.Markets[0].MarketId
	got, err := client.GetMarket(context.Background(), &spotv1.GetMarketRequest{
		MarketId: marketID,
	})
	require.NoError(t, err)
	require.Equal(t, marketID, got.Market.MarketId)
}

func TestE2E_CreateMarket_Forbidden(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"USER"})
	defer cleanup()

	ctx := authCtx(t, "user-1", []string{"USER"})

	_, err := client.CreateMarket(ctx, &spotv1.CreateMarketRequest{
		Name:          "TEST/USDT",
		RequiredRoles: []userv1.Role{userv1.Role_ROLE_USER},
		Enabled:       true,
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestE2E_CreateMarket_Success(t *testing.T) {
	client, cleanup := setupTestServer(t, []string{"ADMIN"})
	defer cleanup()

	ctx := authCtx(t, "admin-1", []string{"ADMIN"})

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

	ctx := authCtx(t, "user-1", []string{"USER"})
	_, err := client.DeleteMarket(ctx, &spotv1.DeleteMarketRequest{
		MarketId: "00000000-0000-0000-0000-000000000000",
	})
	require.Error(t, err)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}
