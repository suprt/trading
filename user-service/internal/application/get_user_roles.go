package application

import (
	"context"

	"github.com/suprt/trading/user-service/internal/core/dto"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
)

type GetUserRolesUseCase struct {
	users outbound.UserRepository
}

func NewGetUserRolesUseCase(users outbound.UserRepository) *GetUserRolesUseCase {
	return &GetUserRolesUseCase{users: users}
}

func (uc *GetUserRolesUseCase) GetUserRoles(ctx context.Context, in dto.GetUserRolesInput) (*dto.GetUserRolesOutput, error) {
	user, err := uc.users.GetByID(ctx, in.UserID)
	if err != nil {
		return nil, err
	}
	return &dto.GetUserRolesOutput{
		Roles:        user.Roles,
		RolesVersion: user.RolesVersion,
	}, nil
}
