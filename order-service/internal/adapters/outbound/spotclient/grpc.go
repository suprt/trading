package spotclient

import (
	"context"

	"github.com/suprt/trading/order-service/internal/adapters/outbound/rolemapper"
	"github.com/suprt/trading/order-service/internal/core/domain"
	er "github.com/suprt/trading/order-service/internal/core/errors"
	spotv1 "github.com/suprt/trading/proto/spot/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SpotClient struct {
	client spotv1.SpotInstrumentServiceClient
}

func NewSpotClient(client spotv1.SpotInstrumentServiceClient) *SpotClient {
	return &SpotClient{client: client}
}

func (c *SpotClient) GetMarket(ctx context.Context, marketID domain.MarketID) (*domain.Market, error) {
	resp, err := c.client.GetMarket(ctx, &spotv1.GetMarketRequest{
		MarketId: string(marketID),
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, er.ErrMarketNotFound
		}
		return nil, err
	}

	roles, err := rolemapper.RolesFromProto(resp.Market.GetRequiredRoles())
	if err != nil {
		return nil, err
	}

	return &domain.Market{
		ID:            domain.MarketID(resp.Market.GetMarketId()),
		Name:          resp.Market.GetName(),
		RequiredRoles: roles,
	}, nil
}
