package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func setupCreateMarketTest() (*mockMarketRepo, *application.CreateMarketUseCase) {
	repo := newMockMarketRepo()
	ids := &mockIDGen{id: "new-id"}
	uc := application.NewCreateMarketUseCase(repo, ids) // без users
	return repo, uc
}

func TestCreateMarket_Success(t *testing.T) {
	repo, uc := setupCreateMarketTest()
	ctx := ctxWithAdmin("admin-1")

	out, err := uc.CreateMarket(ctx, dto.CreateMarketInput{
		Name:          "NEW/USDT",
		RequiredRoles: []domain.Role{domain.RoleUser},
		Enabled:       true,
	})
	require.NoError(t, err)
	require.Equal(t, "new-id", string(out.Market.ID))
	require.Equal(t, "NEW/USDT", out.Market.Name)

	_, ok := repo.markets["new-id"]
	require.True(t, ok)
}

func TestCreateMarket_Forbidden(t *testing.T) {
	repo, uc := setupCreateMarketTest()
	ctx := ctxWithNonAdmin("user-1")

	_, err := uc.CreateMarket(ctx, dto.CreateMarketInput{
		Name:    "NEW/USDT",
		Enabled: true,
	})
	require.ErrorIs(t, err, er.ErrForbidden)
	require.Empty(t, repo.markets)
}

func TestCreateMarket_NameTaken(t *testing.T) {
	repo, uc := setupCreateMarketTest()
	repo.markets["existing"] = &domain.Market{ID: "existing", Name: "BTC/USDT"}
	ctx := ctxWithAdmin("admin-1")

	_, err := uc.CreateMarket(ctx, dto.CreateMarketInput{
		Name:    "BTC/USDT",
		Enabled: true,
	})
	require.ErrorIs(t, err, er.ErrMarketNameTaken)
}
