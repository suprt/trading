package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/dto"
)

func TestViewMarkets_RoleHierarchy(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "SOL/USDT", RequiredRoles: []domain.Role{domain.RoleTrader}, Enabled: true}
	repo.markets["3"] = &domain.Market{ID: "3", Name: "DOGE/USDT", RequiredRoles: []domain.Role{domain.RoleAdmin}, Enabled: true}

	uc := application.NewViewMarketsUseCase(repo)

	// USER
	ctx := auth.WithRoles(context.Background(), []domain.Role{domain.RoleUser})
	out, err := uc.ViewMarkets(ctx, dto.ViewMarketsInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT"}, marketNames(out.Markets))

	// TRADER
	ctx = auth.WithRoles(context.Background(), []domain.Role{domain.RoleTrader})
	out, err = uc.ViewMarkets(ctx, dto.ViewMarketsInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT", "SOL/USDT"}, marketNames(out.Markets))

	// ADMIN
	ctx = auth.WithRoles(context.Background(), []domain.Role{domain.RoleAdmin})
	out, err = uc.ViewMarkets(ctx, dto.ViewMarketsInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT", "SOL/USDT", "DOGE/USDT"}, marketNames(out.Markets))
}

func TestViewMarkets_DisabledAndDeleted(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "ACTIVE", Enabled: true}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "DISABLED", Enabled: false}

	uc := application.NewViewMarketsUseCase(repo)
	ctx := auth.WithRoles(context.Background(), []domain.Role{domain.RoleUser})
	out, err := uc.ViewMarkets(ctx, dto.ViewMarketsInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE"}, marketNames(out.Markets))
}

func TestViewMarkets_Deleted(t *testing.T) {
	repo := newMockMarketRepo()
	now := time.Now()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "DELETED", Enabled: true, DeletedAt: &now}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "ACTIVE", Enabled: true}

	uc := application.NewViewMarketsUseCase(repo)
	ctx := auth.WithRoles(context.Background(), []domain.Role{domain.RoleUser})
	out, err := uc.ViewMarkets(ctx, dto.ViewMarketsInput{})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE"}, marketNames(out.Markets))
}
func TestViewMarkets_NoRoles(t *testing.T) {
	repo := newMockMarketRepo()
	uc := application.NewViewMarketsUseCase(repo)

	_, err := uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{})
	require.ErrorIs(t, err, er.ErrUnauthenticated)
}
