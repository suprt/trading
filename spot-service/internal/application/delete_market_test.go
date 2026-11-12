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

func TestDeleteMarket_Success(t *testing.T) {
	repo := newMockMarketRepo()
	repo.markets["1"] = &domain.Market{ID: "1", Name: "BTC/USDT", Enabled: true}
	users := &mockUserClient{roles: []domain.Role{domain.RoleAdmin}}

	uc := application.NewDeleteMarketUseCase(repo, users)

	err := uc.DeleteMarket(context.Background(), dto.DeleteMarketInput{
		CallerID: "admin-1",
		ID:       "1",
	})
	require.NoError(t, err)

	m := repo.markets["1"]
	require.NotNil(t, m.DeletedAt) // soft-delete сработал
}

func TestDeleteMarket_Forbidden(t *testing.T) {
	repo := newMockMarketRepo()
	users := &mockUserClient{roles: []domain.Role{domain.RoleUser}} // не админ

	uc := application.NewDeleteMarketUseCase(repo, users)

	err := uc.DeleteMarket(context.Background(), dto.DeleteMarketInput{
		CallerID: "user-1",
		ID:       "1",
	})
	require.ErrorIs(t, err, er.ErrForbidden)
}
