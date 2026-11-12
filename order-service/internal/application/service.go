package application

type Service struct {
	*CreateOrderUseCase
	*GetOrderUseCase
	*GetOrderStatusUseCase
	*AdminGetOrderUseCase
	*AdminGetOrderStatusUseCase
	*CancelOrderUseCase
	*AdminCancelOrderUseCase
	*ListOrdersUseCase
	*AdminListOrdersUseCase
}

func NewService(
	createOrder *CreateOrderUseCase,
	getOrder *GetOrderUseCase,
	getOrderStatus *GetOrderStatusUseCase,
	adminGetOrder *AdminGetOrderUseCase,
	adminGetOrderStatus *AdminGetOrderStatusUseCase,
	cancelOrder *CancelOrderUseCase,
	adminCancelOrder *AdminCancelOrderUseCase,
	listOrders *ListOrdersUseCase,
	adminListOrders *AdminListOrdersUseCase,
) *Service {
	return &Service{
		CreateOrderUseCase:         createOrder,
		GetOrderUseCase:            getOrder,
		GetOrderStatusUseCase:      getOrderStatus,
		AdminGetOrderUseCase:       adminGetOrder,
		AdminGetOrderStatusUseCase: adminGetOrderStatus,
		CancelOrderUseCase:         cancelOrder,
		AdminCancelOrderUseCase:    adminCancelOrder,
		ListOrdersUseCase:          listOrders,
		AdminListOrdersUseCase:     adminListOrders,
	}
}
