CREATE TABLE IF NOT EXISTS crates (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    category    TEXT NOT NULL,
    description TEXT NOT NULL,
    logo        TEXT NOT NULL DEFAULT '',
    tags        JSONB NOT NULL DEFAULT '[]',
    official    BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS blueprints (
    id               TEXT PRIMARY KEY,
    crate_id         TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL,
    long_description TEXT NOT NULL DEFAULT '',
    logo             TEXT NOT NULL DEFAULT '',
    image            TEXT NOT NULL,
    version          TEXT NOT NULL,
    official         BOOLEAN NOT NULL DEFAULT false,
    category         TEXT NOT NULL,
    runtime_hints    JSONB NOT NULL DEFAULT '{}',
    resources        JSONB NOT NULL DEFAULT '{}',
    ports            JSONB NOT NULL DEFAULT '[]',
    config           JSONB NOT NULL DEFAULT '[]',
    outputs          JSONB NOT NULL DEFAULT '[]',
    extensions       JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
