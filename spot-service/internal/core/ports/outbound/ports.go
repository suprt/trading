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

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type IDGenerator interface {
	New() string
}

type UserClient interface {
	GetUserRoles(ctx context.Context, userID domain.UserID) ([]domain.Role, int, error)
}

type TokenIssuer interface {
	ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, int64, error)
}

type MarketCache interface {
	GetByID(ctx context.Context, id domain.MarketID) (*domain.Market, error)
	SetByID(ctx context.Context, m *domain.Market) error
	InvalidateByID(ctx context.Context, id domain.MarketID) error

	GetList(ctx context.Context) ([]*domain.Market, error)
	SetList(ctx context.Context, markets []*domain.Market) error
	InvalidateList(ctx context.Context) error

	AcquireLock(ctx context.Context) (bool, error)
	IsLocked(ctx context.Context) (bool, error)
	ReleaseLock(ctx context.Context) error
}

type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}
