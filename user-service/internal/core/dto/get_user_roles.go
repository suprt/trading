package dto

import "github.com/suprt/trading/user-service/internal/core/domain"

type GetUserRolesInput struct {
	UserID domain.UserID
}

type GetUserRolesOutput struct {
	Roles        []domain.Role
	RolesVersion int64
}
