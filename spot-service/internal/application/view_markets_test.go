package application_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
)

func TestViewMarkets_RoleHierarchy(t *testing.T) {
	repo := newMockMarketRepo()
	// BTC - для всех
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	// SOL - для трейдеров
	repo.markets["2"] = &domain.Market{ID: "2", Name: "SOL/USDT", RequiredRoles: []domain.Role{domain.RoleTrader}, Enabled: true}
	// DOGE - для админов
	repo.markets["3"] = &domain.Market{ID: "3", Name: "DOGE/USDT", RequiredRoles: []domain.Role{domain.RoleAdmin}, Enabled: true}

	uc := application.NewViewMarketsUseCase(repo)

	// 1. USER видит только BTC
	out, err := uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{
		UserRoles: []domain.Role{domain.RoleUser},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT"}, marketNames(out.Markets))

	// 2. TRADER видит BTC и SOL
	out, err = uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{
		UserRoles: []domain.Role{domain.RoleTrader},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT", "SOL/USDT"}, marketNames(out.Markets))

	// 3. ADMIN видит всё
	out, err = uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{
		UserRoles: []domain.Role{domain.RoleAdmin},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"BTC/USDT", "SOL/USDT", "DOGE/USDT"}, marketNames(out.Markets))
}

func TestViewMarkets_DisabledAndDeleted(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "ACTIVE", Enabled: true}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "DISABLED", Enabled: false}

	uc := application.NewViewMarketsUseCase(repo)

	out, err := uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{
		UserRoles: []domain.Role{domain.RoleUser},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE"}, marketNames(out.Markets))
}

func TestViewMarkets_Deleted(t *testing.T) {
	repo := newMockMarketRepo()
	now := time.Now()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "DELETED", Enabled: true, DeletedAt: &now}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "ACTIVE", Enabled: true}

	uc := application.NewViewMarketsUseCase(repo)

	out, err := uc.ViewMarkets(context.Background(), dto.ViewMarketsInput{
		UserRoles: []domain.Role{domain.RoleUser},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE"}, marketNames(out.Markets))
}
