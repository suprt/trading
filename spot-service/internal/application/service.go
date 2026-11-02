package application

type Service struct {
	*CreateMarketUseCase
	*DeleteMarketUseCase
	*GetMarketUseCase
	*UpdateMarketDetailsUseCase
	*UpdateMarketRolesUseCase
	*ViewMarketsUseCase
	*ListAllMarketsUseCase
}

func NewService(
	createMarket *CreateMarketUseCase,
	deleteMarket *DeleteMarketUseCase,
	getMarket *GetMarketUseCase,
	updateMarketDetails *UpdateMarketDetailsUseCase,
	updateMarketRoles *UpdateMarketRolesUseCase,
	viewMarkets *ViewMarketsUseCase,
	listAllMarkets *ListAllMarketsUseCase,
) *Service {
	return &Service{
		CreateMarketUseCase:        createMarket,
		DeleteMarketUseCase:        deleteMarket,
		GetMarketUseCase:           getMarket,
		UpdateMarketDetailsUseCase: updateMarketDetails,
		UpdateMarketRolesUseCase:   updateMarketRoles,
		ViewMarketsUseCase:         viewMarkets,
		ListAllMarketsUseCase:      listAllMarkets,
	}
}
