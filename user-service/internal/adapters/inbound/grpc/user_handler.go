package grpc

import (
	"context"

	userv1 "github.com/suprt/trading/proto/user/v1"
	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/dto"
	"github.com/suprt/trading/user-service/internal/core/ports/inbound"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	userv1.UnimplementedUserServiceServer
	userService inbound.UserService
}

func New(userService inbound.UserService) *Handler {
	return &Handler{userService: userService}
}

func (h *Handler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	var idemKey string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("idempotency-key"); len(vals) > 0 {
			idemKey = vals[0]
		}
	}

	out, err := h.userService.Register(ctx, dto.RegisterInput{
		Email:          req.GetEmail(),
		Username:       req.GetUsername(),
		Password:       req.GetPassword(),
		IdempotencyKey: idemKey,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.RegisterResponse{
		UserId: string(out.UserID),
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	out, err := h.userService.Login(ctx, dto.LoginInput{
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.LoginResponse{
		UserId:          string(out.UserID),
		AccessToken:     out.AccessToken.String(),
		RefreshToken:    out.RefreshToken.String(),
		AccessExpiresIn: out.AccessExpiresIn,
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	out, err := h.userService.RefreshToken(ctx, dto.RefreshTokenInput{
		RefreshToken: domain.RefreshToken(req.GetRefreshToken()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.RefreshTokenResponse{
		AccessToken:     out.AccessToken.String(),
		RefreshToken:    out.RefreshToken.String(),
		AccessExpiresIn: out.AccessExpiresIn,
	}, nil
}

func (h *Handler) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	err := h.userService.Logout(ctx, dto.LogoutInput{
		RefreshToken: domain.RefreshToken(req.GetRefreshToken()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.LogoutResponse{}, nil
}

func (h *Handler) GetSelfProfile(ctx context.Context, _ *userv1.GetSelfProfileRequest) (*userv1.GetSelfProfileResponse, error) {
	out, err := h.userService.GetSelfProfile(ctx, dto.GetSelfProfileInput{})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.GetSelfProfileResponse{
		User: userToProto(out.User),
	}, nil
}

func (h *Handler) GetUserByID(ctx context.Context, req *userv1.GetUserByIDRequest) (*userv1.GetUserByIDResponse, error) {
	out, err := h.userService.GetUserByID(ctx, dto.GetUserByIDInput{
		UserID: domain.UserID(req.GetUserId()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.GetUserByIDResponse{
		User: userToProto(out.User),
	}, nil
}

func (h *Handler) AddRole(ctx context.Context, req *userv1.AddRoleRequest) (*userv1.AddRoleResponse, error) {
	role, err := roleFromProto(req.GetRole())
	if err != nil {
		return nil, mapErr(err)
	}
	if err := h.userService.AddRole(ctx, dto.AddRoleInput{
		UserID: domain.UserID(req.GetUserId()),
		Role:   role,
	}); err != nil {
		return nil, mapErr(err)
	}
	return &userv1.AddRoleResponse{}, nil
}

func (h *Handler) RemoveRole(ctx context.Context, req *userv1.RemoveRoleRequest) (*userv1.RemoveRoleResponse, error) {
	role, err := roleFromProto(req.GetRole())
	if err != nil {
		return nil, mapErr(err)
	}
	if err := h.userService.RemoveRole(ctx, dto.RemoveRoleInput{
		UserID: domain.UserID(req.GetUserId()),
		Role:   role,
	}); err != nil {
		return nil, mapErr(err)
	}
	return &userv1.RemoveRoleResponse{}, nil
}

func (h *Handler) GetPublicProfile(ctx context.Context, req *userv1.GetPublicProfileRequest) (*userv1.GetPublicProfileResponse, error) {
	out, err := h.userService.GetPublicProfile(ctx, dto.GetPublicProfileInput{
		UserID: domain.UserID(req.GetUserId()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.GetPublicProfileResponse{
		Profile: &userv1.PublicProfile{
			Id:        string(out.UserID),
			Username:  out.Username,
			CreatedAt: timestamppb.New(out.CreatedAt),
		},
	}, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	in := dto.UpdateProfileInput{}
	if req.Email != nil {
		in.Email = req.Email
	}
	if req.Username != nil {
		in.Username = req.Username
	}

	if err := h.userService.UpdateProfile(ctx, in); err != nil {
		return nil, mapErr(err)
	}
	return &userv1.UpdateProfileResponse{}, nil
}

func (h *Handler) ChangePassword(ctx context.Context, req *userv1.ChangePasswordRequest) (*userv1.ChangePasswordResponse, error) {
	err := h.userService.ChangePassword(ctx, dto.ChangePasswordInput{
		CurrentPassword: req.GetCurrentPassword(),
		NewPassword:     req.GetNewPassword(),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.ChangePasswordResponse{}, nil
}

func (h *Handler) AdminUpdateProfile(ctx context.Context,
	req *userv1.AdminUpdateProfileRequest) (*userv1.AdminUpdateProfileResponse, error) {
	in := dto.AdminUpdateProfileInput{UserID: domain.UserID(req.GetUserId())}
	if req.Email != nil {
		in.Email = req.Email
	}
	if req.Username != nil {
		in.Username = req.Username
	}

	if err := h.userService.AdminUpdateProfile(ctx, in); err != nil {
		return nil, mapErr(err)
	}
	return &userv1.AdminUpdateProfileResponse{}, nil
}

func (h *Handler) AdminResetPassword(ctx context.Context,
	req *userv1.AdminResetPasswordRequest) (*userv1.AdminResetPasswordResponse, error) {
	out, err := h.userService.AdminResetPassword(ctx, dto.AdminResetPasswordInput{
		UserID: domain.UserID(req.GetUserId()),
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.AdminResetPasswordResponse{
		TemporaryPassword: out.TemporaryPassword,
	}, nil

}

func (h *Handler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	out, err := h.userService.ListUsers(ctx, dto.ListUsersInput{
		PageSize:  req.GetPageSize(),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	users := make([]*userv1.User, len(out.Users))
	for i, user := range out.Users {
		users[i] = userToProto(user)
	}

	return &userv1.ListUsersResponse{
		Users:         users,
		NextPageToken: out.NextPageToken,
	}, nil
}

func (h *Handler) GetUserRoles(ctx context.Context, req *userv1.GetUserRolesRequest) (*userv1.GetUserRolesResponse, error) {
	out, err := h.userService.GetUserRoles(ctx, dto.GetUserRolesInput{
		UserID: domain.UserID(req.UserId),
	})
	if err != nil {
		return nil, mapErr(err)
	}

	roles := make([]userv1.Role, len(out.Roles))
	for i, r := range out.Roles {
		roles[i] = roleToProto(r)
	}

	return &userv1.GetUserRolesResponse{
		Roles:        roles,
		RolesVersion: int32(out.RolesVersion),
	}, nil
}
