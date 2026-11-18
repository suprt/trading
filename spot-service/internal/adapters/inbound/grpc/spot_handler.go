package grpc

import (
	"context"

	spotv1 "github.com/suprt/trading/proto/spot/v1"
	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/dto"
	"github.com/suprt/trading/spot-service/internal/core/ports/inbound"
)

type Handler struct {
	spotv1.UnimplementedSpotInstrumentServiceServer
	spotService inbound.SpotService
}

func New(spotService inbound.SpotService) *Handler {
	return &Handler{spotService: spotService}
}

// TODO: ViewMarkets - отображатает доступные для юзера маркеты в зависимоти от ЕГО ролей
func (h *Handler) ViewMarkets(ctx context.Context, req *spotv1.ViewMarketsRequest) (*spotv1.ViewMarketsResponse, error) {

	out, err := h.spotService.ViewMarkets(ctx, dto.ViewMarketsInput{})
	if err != nil {
		return nil, mapError(err)
	}

	markets := make([]*spotv1.Market, len(out.Markets))
	for i, market := range out.Markets {
		markets[i] = marketToProto(market)
	}
	return &spotv1.ViewMarketsResponse{Markets: markets}, nil
}

func (h *Handler) GetMarket(ctx context.Context, req *spotv1.GetMarketRequest) (*spotv1.GetMarketResponse, error) {
	out, err := h.spotService.GetMarket(ctx, dto.GetMarketInput{
		ID: domain.MarketID(req.GetMarketId()),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &spotv1.GetMarketResponse{Market: marketToProto(out.Market)}, nil
}

func (h *Handler) CreateMarket(ctx context.Context, req *spotv1.CreateMarketRequest) (*spotv1.CreateMarketResponse, error) {
	roles, err := rolesFromProto(req.GetRequiredRoles())
	if err != nil {
		return nil, mapError(err)
	}
	out, err := h.spotService.CreateMarket(ctx, dto.CreateMarketInput{
		Name:          req.GetName(),
		RequiredRoles: roles,
		Enabled:       req.GetEnabled(),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &spotv1.CreateMarketResponse{Market: marketToProto(out.Market)}, nil
}

func (h *Handler) UpdateMarketDetails(ctx context.Context,
	req *spotv1.UpdateMarketDetailsRequest) (*spotv1.UpdateMarketDetailsResponse, error) {

	out, err := h.spotService.UpdateMarketDetails(ctx, dto.UpdateMarketDetailsInput{
		ID:      domain.MarketID(req.GetMarketId()),
		Name:    req.Name,
		Enabled: req.Enabled,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &spotv1.UpdateMarketDetailsResponse{Market: marketToProto(out.Market)}, nil
}

func (h *Handler) UpdateMarketRoles(ctx context.Context,
	req *spotv1.UpdateMarketRolesRequest) (*spotv1.UpdateMarketRolesResponse, error) {
	roles, err := rolesFromProto(req.RequiredRoles)
	if err != nil {
		return nil, mapError(err)
	}
	out, err := h.spotService.UpdateMarketRoles(ctx, dto.UpdateMarketRolesInput{
		ID:            domain.MarketID(req.MarketId),
		RequiredRoles: roles,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &spotv1.UpdateMarketRolesResponse{Market: marketToProto(out.Market)}, nil
}

func (h *Handler) DeleteMarket(ctx context.Context, req *spotv1.DeleteMarketRequest) (*spotv1.DeleteMarketResponse, error) {

	err := h.spotService.DeleteMarket(ctx, dto.DeleteMarketInput{
		ID: domain.MarketID(req.MarketId),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &spotv1.DeleteMarketResponse{}, nil
}

func (h *Handler) ListAllMarkets(ctx context.Context, req *spotv1.ListAllMarketsRequest) (*spotv1.ListAllMarketsResponse, error) {

	out, err := h.spotService.ListAllMarkets(ctx, dto.ListAllMarketsInput{
		IncludeDeleted:  req.IncludeDeleted,
		IncludeDisabled: req.IncludeDisabled,
	})
	if err != nil {
		return nil, mapError(err)
	}

	markets := make([]*spotv1.Market, len(out.Markets))
	for i, m := range out.Markets {
		markets[i] = marketToProto(m)
	}
	return &spotv1.ListAllMarketsResponse{Markets: markets}, nil
}
