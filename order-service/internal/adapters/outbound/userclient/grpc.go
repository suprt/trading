package userclient

import (
	"context"

	"google.golang.org/grpc/metadata"

	"github.com/suprt/trading/order-service/internal/adapters/outbound/rolemapper"
	"github.com/suprt/trading/order-service/internal/core/domain"
	userv1 "github.com/suprt/trading/proto/user/v1"
)

type UserClient struct {
	client userv1.UserServiceClient
	secret string
}

func NewUserClient(client userv1.UserServiceClient, secret string) *UserClient {
	return &UserClient{client: client, secret: secret}
}

func (c *UserClient) GetUserRoles(ctx context.Context, userID domain.UserID) ([]domain.Role, int64, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-internal-secret", c.secret)

	resp, err := c.client.GetUserRoles(ctx, &userv1.GetUserRolesRequest{
		UserId: string(userID),
	})
	if err != nil {
		return nil, 0, err
	}

	roles, err := rolemapper.RolesFromProto(resp.Roles)
	if err != nil {
		return nil, 0, err
	}
	return roles, resp.RolesVersion, nil
}
