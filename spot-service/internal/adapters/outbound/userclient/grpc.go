package userclient

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"

	userv1 "github.com/suprt/trading/proto/user/v1"
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type UserClient struct {
	client userv1.UserServiceClient
	secret string
}

func NewUserClient(client userv1.UserServiceClient, secret string) *UserClient {
	return &UserClient{client: client, secret: secret}
}

func (c *UserClient) GetUserRoles(ctx context.Context, userID domain.UserID) ([]domain.Role, int, error) {
	ctx = metadata.AppendToOutgoingContext(ctx, "x-internal-secret", c.secret)

	resp, err := c.client.GetUserRoles(ctx, &userv1.GetUserRolesRequest{
		UserId: string(userID),
	})
	if err != nil {
		return nil, 0, err
	}

	roles := make([]domain.Role, len(resp.Roles))
	for i, r := range resp.Roles {
		roles[i] = domain.Role(strings.TrimPrefix(r.String(), "ROLE_"))
	}
	return roles, int(resp.RolesVersion), nil
}
