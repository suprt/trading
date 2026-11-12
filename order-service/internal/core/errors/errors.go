package errors

import "errors"

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrOrderNotCancellable = errors.New("order is not cancellable")
	ErrInvalidPrice        = errors.New("invalid price")
	ErrInvalidOrderType    = errors.New("invalid order type")
	ErrInvalidOrderSide    = errors.New("invalid order side")
	ErrMarketNotFound      = errors.New("market not found")
	ErrForbidden           = errors.New("forbidden")
	ErrUnauthenticated     = errors.New("unauthenticated")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidRole         = errors.New("invalid role")
	ErrInvalidPageToken    = errors.New("invalid page token")
)
