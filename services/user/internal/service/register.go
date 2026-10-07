package service

import (
	"context"
	"errors"
	"time"

	"github.com/suprt/ITK-proj/services/user/internal/domain"
)

type RegisterConfig struct {
	RefreshTTL time.Duration
}
type RegisterInput struct {
	Email    string
	Password string
}

type RegisterOutput struct {
	UserID       domain.UserID
	AccessToken  domain.AccessToken
	RefreshToken domain.RefreshToken
}

type Register struct {
	users       UserRepository
	credentials CredentialsRepository
	sessions    SessionRepository
	hasher      PasswordHasher
	tokens      TokenIssuer
	ids         IDGenerator
	refreshTTL  time.Duration
}

func NewRegister(
	users UserRepository,
	credentials CredentialsRepository,
	sessions SessionRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	ids IDGenerator,
	cfg RegisterConfig,
) *Register {
	return &Register{
		users:       users,
		credentials: credentials,
		sessions:    sessions,
		hasher:      hasher,
		tokens:      tokens,
		ids:         ids,
		refreshTTL:  cfg.RefreshTTL,
	}
}

func (r *Register) Execute(ctx context.Context, in RegisterInput) (*RegisterOutput, error) {
	if len(in.Password) < 8 {
		return nil, domain.ErrWeakPassword
	}

	_, err := r.users.GetByEmail(ctx, in.Email)
	if err == nil {
		return nil, domain.ErrEmailTaken
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}

	hash, err := r.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        domain.UserID(r.ids.New()),
		Email:     in.Email,
		Roles:     []domain.Role{domain.RoleUser}, //всегда создаётся с User, другие роли могут быть выданы через GrantRole
		CreatedAt: time.Now(),
	}

	// UUIDv4 коллизия крайне маловероятна, но н невозможна
	if err := r.createUserWithRetry(ctx, user); err != nil {
		return nil, err
	}

	creds := &domain.Credentials{UserID: user.ID, PasswordHash: hash}
	if err := r.credentials.Save(ctx, creds); err != nil {
		return nil, err
	}

	access, err := r.tokens.IssueAccess(user.ID, user.Roles)
	if err != nil {
		return nil, err
	}
	refresh, err := r.tokens.NewRefresh()
	if err != nil {
		return nil, err
	}

	session := &domain.RefreshSession{
		Token:     refresh,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(r.refreshTTL),
	}

	if err := r.sessions.Save(ctx, session); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		UserID:       user.ID,
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}

const maxIDAttempts = 3

func (r *Register) createUserWithRetry(ctx context.Context, u *domain.User) error {
	var err error
	for range maxIDAttempts {
		u.ID = domain.UserID(r.ids.New())
		err = r.users.Create(ctx, u)
		if err == nil {
			return nil
		}
		if !errors.Is(err, domain.ErrDuplicateID) {
			return err
		}
	}
	return err
}
