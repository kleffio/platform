-- Migration 003: refactor catalog to Blueprint/Construct split.
--
-- Blueprints are now user-facing only (config, resources, extensions sources).
-- Constructs hold the technical recipe (image, env, ports, runtime_hints, outputs).
-- All old hardcoded seed data is removed; the crate registry adapter repopulates
-- crates, blueprints, and constructs from the remote registry on startup.

-- Remove seeded data so we can restructure cleanly.
DELETE FROM blueprints;
DELETE FROM crates;

-- Drop columns that moved to the constructs table.
ALTER TABLE blueprints
    DROP COLUMN IF EXISTS image,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS long_description,
    DROP COLUMN IF EXISTS runtime_hints,
    DROP COLUMN IF EXISTS ports,
    DROP COLUMN IF EXISTS outputs;

-- Add the link from blueprint → construct.
ALTER TABLE blueprints
    ADD COLUMN IF NOT EXISTS construct_id TEXT NOT NULL DEFAULT '';

-- constructs table — the technical recipe for a blueprint.
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
