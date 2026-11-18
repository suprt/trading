package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suprt/trading/spot-service/internal/application"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func setupUpdateMarketRolesTest() (*mockMarketRepo, *application.UpdateMarketRolesUseCase) {
	repo := newMockMarketRepo()
	uc := application.NewUpdateMarketRolesUseCase(repo) // без users, без ids (если ids не нужен)
	return repo, uc
}

func TestUpdateMarketRoles_Success(t *testing.T) {
	repo, uc := setupUpdateMarketRolesTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithAdmin("admin-1")

	out, err := uc.UpdateMarketRoles(ctx, dto.UpdateMarketRolesInput{
		ID:            "1",
		RequiredRoles: []domain.Role{domain.RoleTrader},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []domain.Role{domain.RoleTrader}, out.Market.RequiredRoles)
}

func TestUpdateMarketRoles_Forbidden(t *testing.T) {
	repo, uc := setupUpdateMarketRolesTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithNonAdmin("user-1")

	_, err := uc.UpdateMarketRoles(ctx, dto.UpdateMarketRolesInput{
		ID:            "1",
		RequiredRoles: []domain.Role{domain.RoleAdmin},
	})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestUpdateMarketRoles_NotFound(t *testing.T) {
	_, uc := setupUpdateMarketRolesTest()
	ctx := ctxWithAdmin("admin-1")

	_, err := uc.UpdateMarketRoles(ctx, dto.UpdateMarketRolesInput{
		ID:            "missing",
		RequiredRoles: []domain.Role{domain.RoleUser},
	})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
}
