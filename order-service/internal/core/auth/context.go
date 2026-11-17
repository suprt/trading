package auth

import (
	"context"

	"github.com/suprt/trading/order-service/internal/core/domain"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
	isAdminKey
)

func WithUserID(ctx context.Context, id domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (domain.UserID, bool) {
	id, ok := ctx.Value(userIDKey).(domain.UserID)
	return id, ok
}

func WithIsAdmin(ctx context.Context, isAdmin bool) context.Context {
	return context.WithValue(ctx, isAdminKey, isAdmin)
}

func IsAdminFromContext(ctx context.Context) bool {
	v, _ := ctx.Value(isAdminKey).(bool)
	return v
}
