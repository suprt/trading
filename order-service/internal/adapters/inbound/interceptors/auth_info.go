package interceptors

import (
	"context"

	"github.com/suprt/trading/order-service/internal/core/auth"
)

// AuthInfo реализует интерфейс AuthInfo для RateLimiter
type AuthInfo struct{}

func NewAuthInfo() *AuthInfo {
	return &AuthInfo{}
}

func (a *AuthInfo) IsAdmin(ctx context.Context) bool {
	return auth.IsAdminFromContext(ctx)
}

func (a *AuthInfo) Category(ctx context.Context) (string, bool) {
	roles, ok := auth.RolesFromContext(ctx)
	if !ok {
		return "", false
	}
	return auth.CategoryFromRoles(roles), true
}