package interceptors

import (
	"context"
	"errors"

	"github.com/suprt/trading/user-service/internal/core/auth"
	"github.com/suprt/trading/user-service/internal/core/domain"
	er "github.com/suprt/trading/user-service/internal/core/errors"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryAdmin(users outbound.UserRepository, methods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, need := methods[info.FullMethod]; !need {
			return handler(ctx, req)
		}
		userID, ok := auth.UserIDFromContext(ctx)
		if !ok {
			return handler(ctx, req)
		}
		user, err := users.GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, er.ErrUserNotFound) {
				return nil, status.Error(codes.Unauthenticated, "user not found")
			}
			return nil, status.Error(codes.Internal, "failed to get user")
		}
		ctx = auth.WithIsAdmin(ctx, user.HasRole(domain.RoleAdmin))
		return handler(ctx, req)
	}
}
