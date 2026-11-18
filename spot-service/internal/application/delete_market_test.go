package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suprt/trading/spot-service/internal/application"

	"github.com/suprt/trading/spot-service/internal/core/domain"

	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func setupDeleteMarketTest() (*mockMarketRepo, *application.DeleteMarketUseCase) {
	repo := newMockMarketRepo()
	uc := application.NewDeleteMarketUseCase(repo) // без users
	return repo, uc
}
func TestDeleteMarket_Success(t *testing.T) {
	repo, uc := setupDeleteMarketTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithAdmin("admin-1")

	err := uc.DeleteMarket(ctx, dto.DeleteMarketInput{ID: "1"})
	require.NoError(t, err)

	m := repo.markets["1"]
	require.NotNil(t, m.DeletedAt)
}

func TestDeleteMarket_Forbidden(t *testing.T) {
	repo, uc := setupDeleteMarketTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	ctx := ctxWithNonAdmin("user-1")

	err := uc.DeleteMarket(ctx, dto.DeleteMarketInput{ID: "1"})
	require.ErrorIs(t, err, er.ErrForbidden)
}

func TestDeleteMarket_NotFound(t *testing.T) {
	_, uc := setupDeleteMarketTest()
	ctx := ctxWithAdmin("admin-1")

	err := uc.DeleteMarket(ctx, dto.DeleteMarketInput{ID: "missing"})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
}
