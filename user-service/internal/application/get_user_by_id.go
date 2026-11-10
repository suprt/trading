package application

import (
	"context"
	"errors"

	"github.com/suprt/trading/user-service/internal/core/auth"
	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/dto"
	er "github.com/suprt/trading/user-service/internal/core/errors"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type GetUserByIDUseCase struct {
	users outbound.UserRepository
}

func NewGetUserByIDUseCase(users outbound.UserRepository) *GetUserByIDUseCase {
	return &GetUserByIDUseCase{users: users}
}

func (uc *GetUserByIDUseCase) GetUserByID(ctx context.Context, in dto.GetUserByIDInput) (*dto.GetUserByIDOutput, error) {
	callerID, ok := auth.UserIDFromContext(ctx)
	if !ok {
		return nil, er.ErrUnauthenticated
	}
	var target *domain.User
	caller, err := uc.users.GetByID(ctx, callerID)
	if err != nil {
		if errors.Is(err, er.ErrUserNotFound) {
			return nil, er.ErrUnauthenticated
		}
		return nil, err
	}
	if !caller.HasRole(domain.RoleAdmin) {
		return nil, er.ErrForbidden
	}

	target, err = uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, err // NotFound - можем сообщить админу, что юзера нет
	}

	return &dto.GetUserByIDOutput{User: target}, nil
}
