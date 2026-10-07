package service

import (
	"context"

	"github.com/suprt/trading/services/user/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type CredentialsRepository interface {
	Save(ctx context.Context, c *domain.Credentials) error
	GetByID(ctx context.Context, id domain.UserID) (*domain.Credentials, error)
}

type SessionRepository interface {
	Save(ctx context.Context, s *domain.RefreshSession) error
	GetByToken(ctx context.Context, token domain.RefreshToken) (*domain.RefreshSession, error)
	Delete(ctx context.Context, t domain.RefreshToken) error
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash string, plain string) error
}

type TokenIssuer interface {
	IssueAccess(userID domain.UserID, roles []domain.Role) (domain.AccessToken, error)
	NewRefresh() (domain.RefreshToken, error)
	ParseAccess(token domain.AccessToken) (domain.UserID, []domain.Role, error)
}

type IDGenerator interface {
	New() string
}
