package handler

import (
	"context"
	"errors"

	userv1 "github.com/suprt/ITK-proj/pkg/proto/user/v1"
	"github.com/suprt/ITK-proj/services/user/internal/domain"
	"github.com/suprt/ITK-proj/services/user/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterService interface {
	Execute(ctx context.Context, in service.RegisterInput) (*service.RegisterOutput, error)
}
type Handler struct {
	userv1.UnimplementedUserServiceServer
	register RegisterService
}

func New(register RegisterService) *Handler {
	return &Handler{register: register}
}

func (h *Handler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	out, err := h.register.Execute(ctx, service.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, mapErr(err)
	}
	return &userv1.RegisterResponse{
		UserId:       string(out.UserID),
		AccessToken:  out.AccessToken.String(),
		RefreshToken: out.RefreshToken.String(),
	}, nil
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
