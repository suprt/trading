package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func setupUpdateMarketDetailsTest() (*mockMarketRepo, *application.UpdateMarketDetailsUseCase) {
	repo := newMockMarketRepo()
	uc := application.NewUpdateMarketDetailsUseCase(repo) // без users, без ids (если ids не нужен)
	return repo, uc
}

func TestUpdateMarketDetails_Success(t *testing.T) {
	repo, uc := setupUpdateMarketDetailsTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithAdmin("admin-1")

	newName := "BTC/USDC"
	out, err := uc.UpdateMarketDetails(ctx, dto.UpdateMarketDetailsInput{
		ID:   "1",
		Name: &newName,
	})
	require.NoError(t, err)
	require.Equal(t, "BTC/USDC", out.Market.Name)
}

func TestUpdateMarketDetails_Forbidden(t *testing.T) {
	repo, uc := setupUpdateMarketDetailsTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithNonAdmin("user-1")

	newName := "HACK"
	_, err := uc.UpdateMarketDetails(ctx, dto.UpdateMarketDetailsInput{ID: "1", Name: &newName})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestUpdateMarketDetails_NotFound(t *testing.T) {
	_, uc := setupUpdateMarketDetailsTest()
	ctx := ctxWithAdmin("admin-1")

	newName := "X"
	_, err := uc.UpdateMarketDetails(ctx, dto.UpdateMarketDetailsInput{ID: "missing", Name: &newName})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
}

func TestUpdateMarketDetails_Disabled(t *testing.T) {
	repo, uc := setupUpdateMarketDetailsTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: false}
	ctx := ctxWithAdmin("admin-1")

	newName := "X"
	_, err := uc.UpdateMarketDetails(ctx, dto.UpdateMarketDetailsInput{ID: "1", Name: &newName})
	require.ErrorIs(t, err, er.ErrMarketNotFound) // !IsActive() → NotFound
}
