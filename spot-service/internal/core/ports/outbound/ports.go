package outbound

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type MarketRepository interface {
	List(ctx context.Context) ([]*domain.Market, error)
	GetByID(ctx context.Context, id domain.MarketID) (*domain.Market, error)
	GetByName(ctx context.Context, name string) (*domain.Market, error)
	Create(ctx context.Context, m *domain.Market) error
	UpdateDetails(ctx context.Context, m *domain.Market) error
	UpdateRoles(ctx context.Context, m *domain.Market) error
	SoftDelete(ctx context.Context, id domain.MarketID, deletedAt time.Time) error
}
