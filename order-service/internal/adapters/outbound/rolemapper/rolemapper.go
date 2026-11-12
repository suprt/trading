package rolemapper

import (
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	userv1 "github.com/suprt/trading/proto/user/v1"
)

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
