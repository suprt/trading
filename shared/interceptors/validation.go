package interceptors

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func UnaryValidation() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		msg, ok := req.(proto.Message)

		if !ok {
			return nil, status.Errorf(codes.InvalidArgument, "invalid request type")
		}
		if err := protovalidate.Validate(msg); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
		}
		return handler(ctx, req)
	}
}
