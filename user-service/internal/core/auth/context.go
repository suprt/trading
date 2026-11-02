package auth

import (
	"context"

	"github.com/suprt/trading/user-service/internal/core/domain"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
)

func WithUserID(ctx context.Context, id domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

func UserIDFromContext(ctx context.Context) (domain.UserID, bool) {
	id, ok := ctx.Value(userIDKey).(domain.UserID)
	return id, ok
}
