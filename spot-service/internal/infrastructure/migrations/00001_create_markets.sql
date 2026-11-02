-- +goose Up
CREATE TABLE markets (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    required_role TEXT[] NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    create_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_markets_active ON markets (enabled, deleted_at) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE markets;