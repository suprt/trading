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

type RemoveRoleUseCase struct {
	users outbound.UserRepository
	ses   outbound.SessionRepository
	uow   outbound.UnitOfWork
}

func NewRemoveRoleUseCase(
	users outbound.UserRepository,
	ses outbound.SessionRepository,
	uow outbound.UnitOfWork,
) *RemoveRoleUseCase {
	return &RemoveRoleUseCase{
		users: users,
		ses:   ses,
		uow:   uow,
	}
}

func (uc *RemoveRoleUseCase) RemoveRole(ctx context.Context, in dto.RemoveRoleInput) error {
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

		if !target.HasRole(in.Role) {
			return nil
		}

		newRoles := make([]domain.Role, 0, len(target.Roles))
		for _, role := range target.Roles {
			if role != in.Role {
				newRoles = append(newRoles, role)
			}
		}

		target.Roles = newRoles
		target.RolesVersion++
		target.UpdatedAt = time.Now()

		if err := uc.users.UpdateRoles(ctx, target); err != nil {
			return err
		}

		//Понижение - принудительный релогин, удаляем рефреш токен
		return uc.ses.DeleteByUserID(ctx, target.ID)
	})

}
