package grpc

import (
	"errors"

	er "github.com/suprt/trading/user-service/internal/core/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapErr(err error) error {
	switch {
	case errors.Is(err, er.ErrEmailTaken),
		errors.Is(err, er.ErrUsernameTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, er.ErrWeakPassword),
		errors.Is(err, er.ErrInvalidPageToken):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, er.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, er.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, er.ErrInvalidToken),
		errors.Is(err, er.ErrExpiredToken),
		errors.Is(err, er.ErrInvalidCredentials),
		errors.Is(err, er.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, er.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
