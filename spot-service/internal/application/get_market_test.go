package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func TestGetMarket_Success(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}

	uc := application.NewGetMarketUseCase(repo)

	out, err := uc.GetMarket(context.Background(), dto.GetMarketInput{ID: "1"})
	require.NoError(t, err)
	require.Equal(t, "BTC/USDT", out.Market.Name)
}

func TestGetMarket_NotFound(t *testing.T) {
	repo := newMockMarketRepo()
	uc := application.NewGetMarketUseCase(repo)

	_, err := uc.GetMarket(context.Background(), dto.GetMarketInput{ID: "missing"})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
}

func TestGetMarket_Disabled(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: false}

	uc := application.NewGetMarketUseCase(repo)

	_, err := uc.GetMarket(context.Background(), dto.GetMarketInput{ID: "1"})
	require.ErrorIs(t, err, er.ErrMarketNotFound)
}
