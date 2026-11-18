package dto

import "github.com/suprt/trading/user-service/internal/core/domain"

type UpdateProfileInput struct {
	UserID   *domain.UserID
	Email    *string
	Username *string
}
