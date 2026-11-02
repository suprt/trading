package interceptors

import (
	"context"
	"strings"

	"github.com/suprt/trading/user-service/internal/core/auth"
	"github.com/suprt/trading/user-service/internal/core/domain"
	"github.com/suprt/trading/user-service/internal/core/ports/outbound"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryAuth(tokens outbound.TokenIssuer, wl map[string]struct{}, internalSecret string) grpc.UnaryServerInterceptor {

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		
		if info.FullMethod == "/user.v1.UserService/GetUserRoles" {
			md, _ := metadata.FromIncomingContext(ctx)
			if vals := md.Get("x-internal-secret"); len(vals) > 0 && vals[0] == internalSecret {
				return handler(ctx, req) 
			}
			return nil, status.Error(codes.PermissionDenied, "internal only")
		}



		if _, exist := wl[info.FullMethod]; exist {
			return handler(ctx, req)
		}

		meta, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := meta.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		raw := strings.TrimPrefix(values[0], "Bearer ")
		if raw == "" {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}

		userID, _, _, err := tokens.ParseAccess(domain.AccessToken(raw))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = auth.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}
