package rediscache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/suprt/trading/spot-service/internal/core/domain"
)

type MarketCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewMarketCache(client *redis.Client, ttl time.Duration) *MarketCache {
	return &MarketCache{client: client, ttl: ttl}
}

func (c *MarketCache) GetByID(ctx context.Context, id domain.MarketID) (*domain.Market, error) {
	data, err := c.client.Get(ctx, "market:"+string(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var m domain.Market
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (c *MarketCache) SetByID(ctx context.Context, m *domain.Market) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "market:"+string(m.ID), data, c.ttl).Err()
}

func (c *MarketCache) InvalidateByID(ctx context.Context, id domain.MarketID) error {
	return c.client.Del(ctx, "market:"+string(id)).Err()
}

func (c *MarketCache) GetList(ctx context.Context) ([]*domain.Market, error) {
	data, err := c.client.Get(ctx, "markets:all").Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var markets []*domain.Market
	if err := json.Unmarshal(data, &markets); err != nil {
		return nil, err
	}
	return markets, nil
}

func (c *MarketCache) SetList(ctx context.Context, markets []*domain.Market) error {
	data, err := json.Marshal(markets)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, "markets:all", data, c.ttl).Err()
}

func (c *MarketCache) InvalidateList(ctx context.Context) error {
	return c.client.Del(ctx, "markets:all").Err()
}

func (c *MarketCache) AcquireLock(ctx context.Context) (bool, error) {
	ok, err := c.client.SetNX(ctx, "markets:lock", "1", 5*time.Second).Result()
	return ok, err
}

func (c *MarketCache) IsLocked(ctx context.Context) (bool, error) {
	n, err := c.client.Exists(ctx, "markets:lock").Result()
	return n > 0, err
}

func (c *MarketCache) ReleaseLock(ctx context.Context) error {
	return c.client.Del(ctx, "markets:lock").Err()
}
