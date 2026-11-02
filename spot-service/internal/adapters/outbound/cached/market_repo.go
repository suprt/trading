package cached

import (
	"context"
	"time"

	"github.com/suprt/trading/spot-service/internal/core/domain"
	"github.com/suprt/trading/spot-service/internal/core/ports/outbound"
)

type MarketRepo struct {
	next     outbound.MarketRepository
	cache    outbound.MarketCache
	reporter outbound.ErrorReporter
}

func NewMarketRepo(next outbound.MarketRepository, cache outbound.MarketCache, reporter outbound.ErrorReporter) *MarketRepo {
	return &MarketRepo{next: next, cache: cache, reporter: reporter}
}

func (r *MarketRepo) List(ctx context.Context) ([]*domain.Market, error) {
	locked, err := r.cache.IsLocked(ctx)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_is_locked"})
		locked = true
	}
	if locked {
		return r.next.List(ctx)
	}

	cached, err := r.cache.GetList(ctx)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_get_list"})
	} else if cached != nil {
		return cached, nil
	}

	markets, err := r.next.List(ctx)
	if err != nil {
		return nil, err
	}
	err = r.cache.SetList(ctx, markets)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_set_list"})
	}

	return markets, nil
}

func (r *MarketRepo) GetByID(ctx context.Context, id domain.MarketID) (*domain.Market, error) {
	locked, err := r.cache.IsLocked(ctx)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_is_locked"})
		locked = true
	}
	if locked {
		return r.next.GetByID(ctx, id)
	}

	cached, err := r.cache.GetByID(ctx, id)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_get_by_id", "market_id": string(id)})
	} else if cached != nil {
		return cached, nil
	}
	markets, err := r.next.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cache.SetByID(ctx, markets); err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_set_by_id", "market_id": string(id)})
	}
	return markets, nil
}

func (r *MarketRepo) GetByName(ctx context.Context, name string) (*domain.Market, error) {
	return r.next.GetByName(ctx, name)
}

func (r *MarketRepo) Create(ctx context.Context, m *domain.Market) error {
	if err := r.next.Create(ctx, m); err != nil {
		return err
	}
	r.invalidate(ctx, m.ID)
	return nil
}

func (r *MarketRepo) UpdateDetails(ctx context.Context, m *domain.Market) error {
	if err := r.next.UpdateDetails(ctx, m); err != nil {
		return err
	}
	r.invalidate(ctx, m.ID)
	return nil
}

func (r *MarketRepo) UpdateRoles(ctx context.Context, m *domain.Market) error {
	if err := r.next.UpdateRoles(ctx, m); err != nil {
		return err
	}
	r.invalidate(ctx, m.ID)
	return nil
}

func (r *MarketRepo) SoftDelete(ctx context.Context, id domain.MarketID, deletedAt time.Time) error {
	if err := r.next.SoftDelete(ctx, id, deletedAt); err != nil {
		return err
	}
	r.invalidate(ctx, id)
	return nil
}

func (r *MarketRepo) invalidate(ctx context.Context, id domain.MarketID) {
	_, err := r.cache.AcquireLock(ctx)
	if err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_acquire_lock"})
	}
	defer func() {
		if err := r.cache.ReleaseLock(ctx); err != nil {
			r.reporter.Report(ctx, err, map[string]string{"op": "cache_release_lock"})
		}
	}()

	if err := r.cache.InvalidateByID(ctx, id); err != nil {
		r.reporter.Report(ctx, err, map[string]string{
			"op":        "cache_invalidate_by_id",
			"market_id": string(id),
		})
	}
	if err := r.cache.InvalidateList(ctx); err != nil {
		r.reporter.Report(ctx, err, map[string]string{"op": "cache_invalidate_list"})
	}
}
