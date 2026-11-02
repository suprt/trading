package errors

import "errors"

var (
	ErrMarketNotFound  = errors.New("market not found")
	ErrMarketNameTaken = errors.New("market name already taken")
)
