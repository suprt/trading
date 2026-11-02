package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	er "github.com/suprt/trading/spot-service/internal/core/errors"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, er.ErrMarketNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, er.ErrMarketNameTaken):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, er.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, er.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
