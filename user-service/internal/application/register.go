package application

import (
	"context"
	"errors"
	"time"

	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/dto"
	er "github.com/suprt/trading/user-service/internal/core/errors"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type RegisterUseCase struct {
	users       outbound.UserRepository
	credentials outbound.CredentialsRepository
	hasher      outbound.PasswordHasher
	ids         outbound.IDGenerator
	uow         outbound.UnitOfWork
	idemStore   outbound.IdempotencyStore
	idemTTL     time.Duration
	reporter    outbound.ErrorReporter
}
type RegisterConfig struct {
	IdempotencyTTL time.Duration
}

func NewRegisterUseCase(
	users outbound.UserRepository,
	credentials outbound.CredentialsRepository,
	hasher outbound.PasswordHasher,
	ids outbound.IDGenerator,
	uow outbound.UnitOfWork,
	idemStore outbound.IdempotencyStore,
	reporter outbound.ErrorReporter,
	cfg RegisterConfig,
) *RegisterUseCase {
	return &RegisterUseCase{
		users:       users,
		credentials: credentials,
		hasher:      hasher,
		ids:         ids,
		uow:         uow,
		idemStore:   idemStore,
		reporter:    reporter,
		idemTTL:     cfg.IdempotencyTTL,
	}
}

func (r *RegisterUseCase) Register(ctx context.Context, in dto.RegisterInput) (*dto.RegisterOutput, error) {

	if in.IdempotencyKey != "" {
		rec, err := r.idemStore.Get(ctx, in.IdempotencyKey)
		if err != nil {
			r.reporter.Report(ctx, err, map[string]string{
				"op":  "idempotency_get",
				"key": in.IdempotencyKey,
			})
		}

		if rec != nil {

			return &dto.RegisterOutput{UserID: domain.UserID(rec.UserID)}, nil
		}
	}

	out, err := r.doRegister(ctx, in)
	if err != nil {
		return nil, err
	}

	if in.IdempotencyKey != "" {
		if err := r.idemStore.Set(ctx, in.IdempotencyKey, &domain.IdempotencyRecord{
			UserID: string(out.UserID),
		}, r.idemTTL); err != nil {
			r.reporter.Report(ctx, err, map[string]string{
				"op":  "idempotency_set",
				"key": in.IdempotencyKey,
			})
		}
	}
	return out, nil
}

func (r *RegisterUseCase) doRegister(ctx context.Context, in dto.RegisterInput) (*dto.RegisterOutput, error) {
	_, err := r.users.GetByEmail(ctx, in.Email)
	if err == nil {
		return nil, er.ErrEmailTaken
	}
	if !errors.Is(err, er.ErrUserNotFound) {
		return nil, err
	}
	hash, err := r.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user := &domain.User{
		ID:           domain.UserID(r.ids.New()),
		Email:        in.Email,
		Username:     in.Username,
		Roles:        []domain.Role{domain.RoleUser}, //всегда создаётся с User, другие роли могут быть выданы через AddRole
		RolesVersion: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	creds := &domain.Credentials{UserID: user.ID, PasswordHash: hash}

	err = r.uow.Do(ctx, func(ctx context.Context) error {
		if err := r.users.Create(ctx, user); err != nil {
			return err
		}
		return r.credentials.Save(ctx, creds)
	})

	if err != nil {
		return nil, err
	}

	return &dto.RegisterOutput{
		UserID: user.ID,
	}, nil
}
