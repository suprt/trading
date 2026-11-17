package userclient

import (
	"context"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"google.golang.org/grpc/metadata"

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

	roles, err := RolesFromProto(resp.Roles)
	if err != nil {
		return nil, 0, err
	}
	//TODO: переделать на switch как в мапере
	return roles, resp.RolesVersion, nil
}

func roleFromProto(r userv1.Role) (domain.Role, error) {
	switch r {
	case userv1.Role_ROLE_USER:
		return domain.RoleUser, nil
	case userv1.Role_ROLE_TRADER:
		return domain.RoleTrader, nil
	case userv1.Role_ROLE_ADMIN:
		return domain.RoleAdmin, nil
	default:
		return "", er.ErrInvalidRole
	}
}

func RolesFromProto(roles []userv1.Role) ([]domain.Role, error) {
	result := make([]domain.Role, 0, len(roles))
	for _, r := range roles {
		mapped, err := roleFromProto(r)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}
