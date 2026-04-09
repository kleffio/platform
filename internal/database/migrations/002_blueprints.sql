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
    id           TEXT PRIMARY KEY,
    crate_id     TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    construct_id TEXT NOT NULL DEFAULT '',
    name         TEXT NOT NULL,
    description  TEXT NOT NULL,
    logo         TEXT NOT NULL DEFAULT '',
    version      TEXT NOT NULL,
    official     BOOLEAN NOT NULL DEFAULT false,
    resources    JSONB NOT NULL DEFAULT '{}',
    config       JSONB NOT NULL DEFAULT '[]',
    extensions   JSONB NOT NULL DEFAULT '{}',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS constructs (
    id            TEXT PRIMARY KEY,
    crate_id      TEXT NOT NULL REFERENCES crates(id) ON DELETE CASCADE,
    blueprint_id  TEXT NOT NULL REFERENCES blueprints(id) ON DELETE CASCADE,
    image         TEXT NOT NULL,
    version       TEXT NOT NULL,
    env           JSONB NOT NULL DEFAULT '{}',
    ports         JSONB NOT NULL DEFAULT '[]',
    runtime_hints JSONB NOT NULL DEFAULT '{}',
    extensions    JSONB NOT NULL DEFAULT '{}',
    outputs       JSONB NOT NULL DEFAULT '[]',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
