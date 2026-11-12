package application_test

import (
	"context"
	"time"

	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
)

// ==========================================
// MockOrderRepo
// ==========================================

type mockOrderRepo struct {
	orders map[domain.OrderID]*domain.Order
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[domain.OrderID]*domain.Order)}
}

func (m *mockOrderRepo) Create(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id domain.OrderID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, er.ErrOrderNotFound
	}
	return o, nil
}

func (m *mockOrderRepo) Cancel(_ context.Context, id domain.OrderID, userID *domain.UserID, now time.Time) (*domain.Order, bool, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, false, nil
	}
	if userID != nil && o.UserID != *userID {
		return nil, false, nil
	}
	if o.Status != domain.OrderStatusCreated {
		return nil, false, nil
	}
	o.Status = domain.OrderStatusCancelled
	o.UpdatedAt = now
	return o, true, nil
}

func (m *mockOrderRepo) ListByUser(_ context.Context, userID domain.UserID, limit int,
	_ time.Time, _ domain.OrderID) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, o := range m.orders {
		if o.UserID == userID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (m *mockOrderRepo) ListAll(_ context.Context, userID *domain.UserID, limit int,
	_ time.Time, _ domain.OrderID) ([]*domain.Order, error) {
	var result []*domain.Order
	for _, o := range m.orders {
		if userID == nil || o.UserID == *userID {
			result = append(result, o)
		}
	}
	return result, nil
}

// ==========================================
// MockUserClient
// ==========================================

type mockUserClient struct {
	roles []domain.Role
	err   error
}

func (m *mockUserClient) GetUserRoles(_ context.Context, _ domain.UserID) ([]domain.Role, int64, error) {
	return m.roles, 1, m.err
}

// ==========================================
// MockSpotClient
// ==========================================

type mockSpotClient struct {
	market *domain.Market
	err    error
}

func (m *mockSpotClient) GetMarket(_ context.Context, _ domain.MarketID) (*domain.Market, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.market, nil
}

// ==========================================
// MockIDGen
// ==========================================

type mockIDGen struct {
	id string
}

func (m *mockIDGen) New() string {
	return m.id
}

// ==========================================
// MockErrorReporter
// ==========================================

type mockErrorReporter struct{}

func (m *mockErrorReporter) Report(_ context.Context, _ error, _ map[string]string) {}

// ==========================================
// Хелперы
// ==========================================

func ptr[T any](v T) *T {
	return &v
}

func orderNames(orders []*domain.Order) []string {
	ids := make([]string, len(orders))
	for i, o := range orders {
		ids[i] = string(o.ID)
	}
	return ids
}
