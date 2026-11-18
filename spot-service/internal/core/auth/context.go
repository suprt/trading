package auth

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	rolesKey
	isAdminKey
)

func WithUserID(ctx context.Context, id domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (domain.UserID, bool) {
	id, ok := ctx.Value(userIDKey).(domain.UserID)
	return id, ok
}

func WithRoles(ctx context.Context, roles []domain.Role) context.Context {
	return context.WithValue(ctx, rolesKey, roles)
}

func RolesFromContext(ctx context.Context) ([]domain.Role, bool) {
	roles, ok := ctx.Value(rolesKey).([]domain.Role)
	return roles, ok
}

func WithIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, isAdminKey, isAdmin)
}

func IsAdminFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(isAdminKey).(bool)
	return v
}
