package outbound

import (
	"context"
	"time"

	"github.com/suprt/trading/user-service/internal/core/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	UpdateProfile(ctx context.Context, u *domain.User) error
	UpdateRoles(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, limit int, afterCreatedAt time.Time, afterID domain.UserID) ([]*domain.User, error)
}

type CredentialsRepository interface {
	Save(ctx context.Context, c *domain.Credentials) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.Credentials, error)
}

type SessionRepository interface {
	Save(ctx context.Context, s *domain.RefreshSession) error
	GetByToken(ctx context.Context, token domain.RefreshToken) (*domain.RefreshSession, error)
	Delete(ctx context.Context, t domain.RefreshToken) error
	DeleteByUserID(ctx context.Context, userID domain.UserID) error
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) error
}

type TokenIssuer interface {
	IssueAccess(userID domain.UserID, roles []domain.Role, rolesVersion int64) (domain.AccessToken, error)
	NewRefresh() (domain.RefreshToken, error)
	ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, int64, error)
}

type IDGenerator interface {
	New() string
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserCache interface {
	Get(ctx context.Context, id domain.UserID) (*domain.User, error)
	Set(ctx context.Context, u *domain.User) error
	SetNotFound(ctx context.Context, id domain.UserID) error
	Delete(ctx context.Context, id domain.UserID) error
	DeleteNotFound(ctx context.Context, id domain.UserID) error
}

// ErrorReporter используется для логирования не критических ошибок,
// которые следует отслеживать, например, ошибка инвалидации кэша и т.п.
type ErrorReporter interface {
	Report(ctx context.Context, err error, attrs map[string]string)
}

type IdempotencyStore interface {
	Get(ctx context.Context, key string) (*domain.IdempotencyRecord, error)
	Set(ctx context.Context, key string, rec *domain.IdempotencyRecord, ttl time.Duration) error
}
