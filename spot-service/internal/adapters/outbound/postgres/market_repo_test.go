package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/suprt/trading/spot-service/internal/adapters/outbound/postgres"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func TestMarketRepo_CreateAndByID(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewMarketRepo(pool)
	ctx := context.Background()

	m := &domain.Market{
		ID:            domain.MarketID(uuid.NewString()),
		Name:          "BTC/USDT",
		RequiredRoles: []domain.Role{domain.RoleUser, domain.RoleTrader},
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	require.NoError(t, repo.Create(ctx, m))

	got, err := repo.GetByID(ctx, m.ID)
	require.NoError(t, err)
	require.Equal(t, m.Name, got.Name)
	require.ElementsMatch(t, m.RequiredRoles, got.RequiredRoles)
}

func TestMarketRepo_Create_NameTaken(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewMarketRepo(pool)
	ctx := context.Background()

	m1 := &domain.Market{ID: domain.MarketID(uuid.NewString()), Name: "BTC/USDT", Enabled: true}
	require.NoError(t, repo.Create(ctx, m1))

	m2 := &domain.Market{ID: domain.MarketID(uuid.NewString()), Name: "BTC/USDT", Enabled: true}
	err := repo.Create(ctx, m2)
	require.ErrorIs(t, err, er.ErrMarketNameTaken)
}

func TestMarketRepo_UpdateRoles(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewMarketRepo(pool)
	ctx := context.Background()

	m := &domain.Market{
		ID:            domain.MarketID(uuid.NewString()),
		Name:          "TEST/USDT",
		RequiredRoles: []domain.Role{domain.RoleUser},
		Enabled:       true,
	}
	require.NoError(t, repo.Create(ctx, m))

	// Обновляем роли
	m.RequiredRoles = []domain.Role{domain.RoleTrader}
	m.UpdatedAt = time.Now()
	require.NoError(t, repo.UpdateRoles(ctx, m))

	got, err := repo.GetByID(ctx, m.ID)
	require.NoError(t, err)
	require.ElementsMatch(t, []domain.Role{domain.RoleTrader}, got.RequiredRoles)
}

func TestMarketRepo_SoftDelete(t *testing.T) {
	pool, cleanup := setupPostgres(t)
	defer cleanup()

	repo := postgres.NewMarketRepo(pool)
	ctx := context.Background()

	m := &domain.Market{ID: domain.MarketID(uuid.NewString()), Name: "BTC/USDT", Enabled: true}
	require.NoError(t, repo.Create(ctx, m))

	now := time.Now()
	require.NoError(t, repo.SoftDelete(ctx, m.ID, now))

	got, err := repo.GetByID(ctx, m.ID)
	require.NoError(t, err)
	require.NotNil(t, got.DeletedAt)

}
