package interceptors

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ctxKey string

const RequestIDKey ctxKey = "x-request-id"

func XRequestID() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var reqID string

		if meta, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := meta.Get("x-request-id"); len(vals) > 0 {
				reqID = vals[0]
			}
		}

		if reqID == "" {
			reqID = uuid.NewString()
		}

		ctx = context.WithValue(ctx, RequestIDKey, reqID)
		//ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", reqID)

		_ = grpc.SetHeader(ctx, metadata.Pairs("x-request-id", reqID))
		return handler(ctx, req)
	}
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(RequestIDKey).(string)
	return id, ok
}
