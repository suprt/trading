package dto

import "github.com/suprt/trading/order-service/internal/core/domain"

type AdminListOrdersInput struct {
	UserID    *domain.UserID // nil = все заказы
	PageSize  int32
	PageToken string
}

type AdminListOrdersOutput struct {
	Orders        []*domain.Order
	NextPageToken string
}
