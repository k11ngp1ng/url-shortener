CREATE TABLE IF NOT EXISTS urls (
    code VARCHAR(32) PRIMARY KEY,
    original_url TEXT NOT NULL,
    clicks BIGINT NOT NULL DEFAULT 0 CHECK (clicks >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS urls_expires_at_idx ON urls (expires_at) WHERE expires_at IS NOT NULL;
