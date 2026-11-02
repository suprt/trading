package errors

import "errors"

var (
	ErrMarketNotFound  = errors.New("market not found")
	ErrMarketNameTaken = errors.New("market name already taken")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidToken    = errors.New("invalid refresh token")
	ErrInvalidRole     = errors.New("invalid role")
)
