package interceptors

import (
	"context"
	"log"
	"slices"

	"github.com/suprt/trading/order-service/internal/core/auth"
	"github.com/suprt/trading/order-service/internal/core/domain"
	"github.com/suprt/trading/order-service/internal/core/ports/outbound"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryAdmin(users outbound.UserClient, methods map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if len(methods) == 0 {
			log.Println("no methods")
		}
		if _, need := methods[info.FullMethod]; !need {
			return handler(ctx, req)
		}

		userID, ok := auth.UserIDFromContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		roles, _, err := users.GetUserRoles(ctx, userID)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get roles")
		}

		ctx = auth.WithIsAdmin(ctx, slices.Contains(roles, domain.RoleAdmin))
		return handler(ctx, req)
	}
}
