package application

import (
	"context"
	"errors"
	"time"

	"github.com/suprt/trading/user-service/internal/core/auth"
	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/dto"
	er "github.com/suprt/trading/user-service/internal/core/errors"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type AddRoleUseCase struct {
	users outbound.UserRepository
	uow   outbound.UnitOfWork
}

func NewAddRoleUseCase(users outbound.UserRepository, uow outbound.UnitOfWork) *AddRoleUseCase {
	return &AddRoleUseCase{users: users, uow: uow}
}

func (uc *AddRoleUseCase) AddRole(ctx context.Context, in dto.AddRoleInput) error {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return er.ErrUnauthenticated
	}

	return uc.uow.Do(ctx, func(ctx context.Context) error {
		caller, err := uc.users.GetByID(ctx, callerID)
		if err != nil {
			if errors.Is(err, er.ErrUserNotFound) {
				return er.ErrUnauthenticated
			}
			return err
		}
		if !caller.HasRole(domain.RoleAdmin) {
			return er.ErrForbidden
		}

		target, err := uc.users.GetByID(ctx, in.UserID)
		if err != nil {
			return err
		}

		if target.HasRole(in.Role) {
			return nil
		}

		target.Roles = append(target.Roles, in.Role)
		target.RolesVersion++
		target.UpdatedAt = time.Now()

		return uc.users.UpdateRoles(ctx, target)
	})
}
