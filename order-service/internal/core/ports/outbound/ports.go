package outbound

import (
	"context"
	"time"

	"github.com/suprt/trading/order-service/internal/core/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, o *domain.Order) error
	GetByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
	// Cancel возвращает (order, true, nil) при успехе, (nil, false, nil) если строка не обновлена.
	Cancel(ctx context.Context, id domain.OrderID, userID *domain.UserID, now time.Time) (*domain.Order, bool, error)
	ListByUser(ctx context.Context, userID domain.UserID, limit int, afterCreatedAt time.Time, afterID domain.OrderID) ([]*domain.Order, error)
	ListAll(ctx context.Context, userID *domain.UserID, limit int, afterCreatedAt time.Time, afterID domain.OrderID) ([]*domain.Order, error)
}

type UserClient interface {
	GetUserRoles(ctx context.Context, userID domain.UserID) ([]domain.Role, int64, error)
}

type SpotClient interface {
	GetMarket(ctx context.Context, marketID domain.MarketID) (*domain.Market, error)
}

type TokenIssuer interface {
	ParseAccess(token domain.AccessToken) (domain.UserID, error)
}

type IDGenerator interface {
	New() string
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}
