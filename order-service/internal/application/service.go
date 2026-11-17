package application

type Service struct {
	*CreateOrderUseCase
	*GetOrderUseCase
	*GetOrderStatusUseCase
	*CancelOrderUseCase
	*ListOrdersUseCase
}

func NewService(
	createOrder *CreateOrderUseCase,
	getOrder *GetOrderUseCase,
	getOrderStatus *GetOrderStatusUseCase,
	cancelOrder *CancelOrderUseCase,
	listOrders *ListOrdersUseCase,
) *Service {
	return &Service{
		CreateOrderUseCase:    createOrder,
		GetOrderUseCase:       getOrder,
		GetOrderStatusUseCase: getOrderStatus,
		CancelOrderUseCase:    cancelOrder,
		ListOrdersUseCase:     listOrders,
	}
}
