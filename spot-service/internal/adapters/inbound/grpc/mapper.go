package grpc

import (
	spotv1 "github.com/suprt/trading/proto/spot/v1"
	userv1 "github.com/suprt/trading/proto/user/v1"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	er "github.com/suprt/trading/spot-service/internal/core/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func marketToProto(market *domain.Market) *spotv1.Market {
	roles := make([]userv1.Role, len(market.RequiredRoles))
	for i, r := range market.RequiredRoles {
		roles[i] = roleToProto(r)
	}

	var deletedAt *timestamppb.Timestamp
	if market.DeletedAt != nil {
		deletedAt = timestamppb.New(*market.DeletedAt)
	}
	return &spotv1.Market{
		MarketId:      string(market.ID),
		Name:          market.Name,
		RequiredRoles: roles,
		Enabled:       market.Enabled,
		CreatedAt:     timestamppb.New(market.CreatedAt),
		UpdatedAt:     timestamppb.New(market.UpdatedAt),
		DeletedAt:     deletedAt,
	}
}

func roleToProto(r domain.Role) userv1.Role {
	switch r {
	case domain.RoleUser:
		return userv1.Role_ROLE_USER
	case domain.RoleTrader:
		return userv1.Role_ROLE_TRADER
	case domain.RoleAdmin:
		return userv1.Role_ROLE_ADMIN
	default:
		return userv1.Role_ROLE_UNSPECIFIED
	}
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

func rolesFromProto(roles []userv1.Role) ([]domain.Role, error) {
	result := make([]domain.Role, 0, len(roles))
	for _, r := range roles {
		mapped, err := roleFromProto(r) //TODO: err если роль не валидна
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}
	return result, nil
}
