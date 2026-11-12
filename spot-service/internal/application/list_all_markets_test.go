package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/suprt/trading/spot-service/internal/core/domain"

	"github.com/suprt/trading/spot-service/internal/application"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func TestListAllMarkets_Success(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "ACTIVE", Enabled: true}
	repo.markets["2"] = &domain.Market{ID: "2", Name: "DISABLED", Enabled: false}
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}

	uc := application.NewListAllMarketsUseCase(repo, users)

	out, err := uc.ListAllMarkets(context.Background(), dto.ListAllMarketsInput{
		CallerID:        "admin-1",
		IncludeDisabled: true,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"ACTIVE", "DISABLED"}, marketNames(out.Markets))
}

func TestListAllMarkets_Forbidden(t *testing.T) {
	repo := newMockMarketRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleUser}}

	uc := application.NewListAllMarketsUseCase(repo, users)

	_, err := uc.ListAllMarkets(context.Background(), dto.ListAllMarketsInput{
		CallerID: "user-1",
	})
	require.ErrorIs(t, err, er.ErrForbidden)
}
