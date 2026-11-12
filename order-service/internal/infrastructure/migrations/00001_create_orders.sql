-- +goose Up
CREATE TABLE orders (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    market_id UUID NOT NULL,
    side TEXT NOT NULL,
    type TEXT NOT NULL,
    price BIGINT NOT NULL DEFAULT 0,
    quantity BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_created ON orders (user_id, created_at DESC, id DESC);
CREATE INDEX idx_orders_created ON orders (created_at DESC, id DESC);

-- +goose Down
DROP TABLE orders;