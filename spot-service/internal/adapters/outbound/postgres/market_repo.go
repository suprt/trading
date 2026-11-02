package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type MarketRepo struct {
	pool *pgxpool.Pool
}

func NewMarketRepo(pool *pgxpool.Pool) *MarketRepo {
	return &MarketRepo{pool: pool}
}
