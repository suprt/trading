-- +goose Up
CREATE TABLE markets (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    required_roles TEXT[] NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT markets_name_unique UNIQUE (name)
);

CREATE INDEX idx_markets_active ON markets (enabled, deleted_at) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE markets;