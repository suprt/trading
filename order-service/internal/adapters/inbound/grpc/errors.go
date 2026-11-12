package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	er "github.com/suprt/trading/order-service/internal/core/errors"
)

func mapError(err error) error {
	switch {
	case errors.Is(err, er.ErrOrderNotFound),
		errors.Is(err, er.ErrMarketNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, er.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, er.ErrUnauthenticated),
		errors.Is(err, er.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, er.ErrInvalidPrice),
		errors.Is(err, er.ErrInvalidOrderType),
		errors.Is(err, er.ErrInvalidRole),
		errors.Is(err, er.ErrInvalidPageToken),
		errors.Is(err, er.ErrInvalidOrderSide):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, er.ErrOrderNotCancellable):
		return status.Error(codes.FailedPrecondition, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}
