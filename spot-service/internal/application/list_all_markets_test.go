package application_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/domain"

	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func setupListAllMarketsTest() (*mockMarketRepo, *application.ListAllMarketsUseCase) {
	repo := newMockMarketRepo()
	uc := application.NewListAllMarketsUseCase(repo) // без users
	return repo, uc
}

func TestListAllMarkets_Success(t *testing.T) {
	repo, uc := setupListAllMarketsTest()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "ACTIVE", Enabled: true}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "DISABLED", Enabled: false}
	ctx := ctxWithAdmin("admin-1")

	out, err := uc.ListAllMarkets(ctx, dto.ListAllMarketsInput{IncludeDisabled: true})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE", "DISABLED"}, marketNames(out.Markets))
}

func TestListAllMarkets_Forbidden(t *testing.T) {
	_, uc := setupListAllMarketsTest()
	ctx := ctxWithNonAdmin("user-1")

	_, err := uc.ListAllMarkets(ctx, dto.ListAllMarketsInput{})
	require.ErrorIs(t, err, er.ErrForbidden)
}
