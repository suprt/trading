package application_test

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/auth"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

// ==========================================
// MockUserClient
// ==========================================

// ==========================================
// MockMarketRepo
// ==========================================

type mockMarketRepo struct {
	markets map[domain.MarketID]*domain.Market
}

func newMockMarketRepo() *mockMarketRepo {
	return &mockMarketRepo{markets: make(map[domain.MarketID]*domain.Market)}
}

func (m *mockMarketRepo) List(_ context.Context) ([]*domain.Market, error) {
	var result []*domain.Market
	for _, mk := range m.markets {
		// В реальном репо List возвращает все (декоратор/юзкейс фильтрует)
		// Но если тестируем use case напрямую (без декоратора) - вернем все
		result = append(result, mk)
	}
	return result, nil
}

func (m *mockMarketRepo) GetByID(_ context.Context, id domain.MarketID) (*domain.Market, error) {
	mk, ok := m.markets[id]
	if !ok {
		return nil, er.ErrMarketNotFound
	}
	return mk, nil
}

func (m *mockMarketRepo) GetByName(_ context.Context, name string) (*domain.Market, error) {
	for _, mk := range m.markets {
		if mk.Name == name {
			return mk, nil
		}
	}
	return nil, er.ErrMarketNotFound
}

func (m *mockMarketRepo) Create(_ context.Context, mk *domain.Market) error {
	for _, existing := range m.markets {
		if existing.Name == mk.Name {
			return er.ErrMarketNameTaken
		}
	}
	m.markets[mk.ID] = mk
	return nil
}

func (m *mockMarketRepo) UpdateDetails(_ context.Context, mk *domain.Market) error {
	if _, ok := m.markets[mk.ID]; !ok {
		return er.ErrMarketNotFound
	}
	for id, existing := range m.markets {
		if existing.Name == mk.Name && id != mk.ID {
			return er.ErrMarketNameTaken
		}
	}
	m.markets[mk.ID] = mk
	return nil
}

func (m *mockMarketRepo) UpdateRoles(_ context.Context, mk *domain.Market) error {
	if _, ok := m.markets[mk.ID]; !ok {
		return er.ErrMarketNotFound
	}
	m.markets[mk.ID] = mk
	return nil
}

func (m *mockMarketRepo) SoftDelete(_ context.Context, id domain.MarketID, deletedAt time.Time) error {
	mk, ok := m.markets[id]
	if !ok || mk.DeletedAt != nil {
		return er.ErrMarketNotFound
	}
	mk.DeletedAt = &deletedAt
	mk.UpdatedAt = deletedAt
	return nil
}

func (m *mockMarketRepo) ListAll(_ context.Context, includeDeleted bool, includeDisabled bool) ([]*domain.Market, error) {
	var result []*domain.Market
	for _, mk := range m.markets {
		if !includeDeleted && mk.DeletedAt != nil {
			continue
		}
		if !includeDisabled && !mk.Enabled {
			continue
		}
		result = append(result, mk)
	}
	return result, nil
}

// ==========================================
// MockIDGenerator
// ==========================================

type mockIDGen struct {
	id string
}

func (m *mockIDGen) New() string {
	return m.id
}

// ==========================================
// Хелперы
// ==========================================

func marketNames(markets []*domain.Market) []string {
	names := make([]string, len(markets))
	for i, m := range markets {
		names[i] = m.Name
	}
	return names
}

func ctxWithUser(id domain.UserID) context.Context {
	return auth.WithUserID(context.Background(), id)
}

func ctxWithAdmin(id domain.UserID) context.Context {
	ctx := auth.WithUserID(context.Background(), id)
	return auth.WithIsAdmin(ctx, true)
}

func ctxWithNonAdmin(id domain.UserID) context.Context {
	ctx := auth.WithUserID(context.Background(), id)
	return auth.WithIsAdmin(ctx, false)
}
