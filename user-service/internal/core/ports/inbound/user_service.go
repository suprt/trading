//go:generate mockgen -source=user_service.go -destination=mocks/mock_user_service.go -package=mocks

package inbound

import (
	"context"

	"github.com/suprt/trading/user-service/internal/core/dto"
)

type UserService interface {
	Register(ctx context.Context, in dto.RegisterInput) (*dto.RegisterOutput, error)
	Login(ctx context.Context, in dto.LoginInput) (*dto.LoginOutput, error)
	RefreshToken(ctx context.Context, in dto.RefreshTokenInput) (*dto.RefreshTokenOutput, error)
	Logout(ctx context.Context, in dto.LogoutInput) error
	GetSelfProfile(ctx context.Context, in dto.GetSelfProfileInput) (*dto.GetSelfProfileOutput, error)
	GetUserByID(ctx context.Context, in dto.GetUserByIDInput) (*dto.GetUserByIDOutput, error)
	AddRole(ctx context.Context, in dto.AddRoleInput) error
	RemoveRole(ctx context.Context, in dto.RemoveRoleInput) error
	GetPublicProfile(ctx context.Context, in dto.GetPublicProfileInput) (*dto.GetPublicProfileOutput, error)
	UpdateProfile(ctx context.Context, in dto.UpdateProfileInput) error
	ChangePassword(ctx context.Context, in dto.ChangePasswordInput) error
	AdminUpdateProfile(ctx context.Context, in dto.AdminUpdateProfileInput) error
	AdminResetPassword(ctx context.Context, in dto.AdminResetPasswordInput) (*dto.AdminResetPasswordOutput, error)
	ListUsers(ctx context.Context, in dto.ListUsersInput) (*dto.ListUsersOutput, error)
	GetUserRoles(ctx context.Context, in dto.GetUserRolesInput) (*dto.GetUserRolesOutput, error)
}
