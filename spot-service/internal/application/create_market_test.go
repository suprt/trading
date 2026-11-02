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

func TestCreateMarket_Success(t *testing.T) {
	repo := newMockMarketRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	ids := &mockIDGen{id: "new-id"}

	uc := application.NewCreateMarketUseCase(repo, users, ids)

	out, err := uc.CreateMarket(context.Background(), dto.CreateMarketInput{
		CallerID:      "admin-1",
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
	repo := newMockMarketRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleUser}} // не админ
	ids := &mockIDGen{id: "new-id"}

	uc := application.NewCreateMarketUseCase(repo, users, ids)

	_, err := uc.CreateMarket(context.Background(), dto.CreateMarketInput{
		CallerID: "user-1",
		Name:     "NEW/USDT",
		Enabled:  true,
	})
	require.ErrorIs(t, err, er.ErrForbidden)
	require.Empty(t, repo.markets)
}

func TestCreateMarket_NameTaken(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["existing"] = &domain.Market{ID: "existing", Name: "BTC/USDT"}
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}
	ids := &mockIDGen{id: "new-id"}

	uc := application.NewCreateMarketUseCase(repo, users, ids)

	_, err := uc.CreateMarket(context.Background(), dto.CreateMarketInput{
		CallerID: "admin-1",
		Name:     "BTC/USDT",
		Enabled:  true,
	})
	require.ErrorIs(t, err, er.ErrMarketNameTaken)
}
