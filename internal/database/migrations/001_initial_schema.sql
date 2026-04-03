CREATE TABLE IF NOT EXISTS plugins (
    id           TEXT PRIMARY KEY,
    type         TEXT NOT NULL,
    display_name TEXT NOT NULL,
    image        TEXT NOT NULL,
    version      TEXT NOT NULL,
    grpc_addr    TEXT NOT NULL DEFAULT '',
    frontend_url TEXT,
    config       JSONB NOT NULL DEFAULT '{}',
    secrets      JSONB NOT NULL DEFAULT '{}',
    enabled      BOOLEAN NOT NULL DEFAULT true,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
