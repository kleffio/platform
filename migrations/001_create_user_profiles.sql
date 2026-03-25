-- Migration: 001_create_user_profiles
-- Creates the user_profiles table, which stores application-level profile data
-- for users authenticated via Ory Kratos (or Hydra OIDC).
--
-- Split Architecture note:
--   Kratos owns the identity record (email, password, MFA, sessions).
--   This table owns everything the application cares about beyond identity:
--   avatar, bio, preferences, etc.
--
-- The `id` column is intentionally set to the Kratos identity.id (the OIDC
-- `sub` claim). No surrogate key is needed — the identity ID IS the profile ID.
-- This makes JOINs trivial and avoids an extra lookup on every request.

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- for gen_random_uuid(), only needed if you want uuid default

CREATE TABLE IF NOT EXISTS user_profiles (
    -- Primary key = Kratos identity.id / OIDC subject claim.
    -- TEXT instead of UUID so the repo stays consistent with the platform's
    -- current string-based ID convention (32-char hex or UUID string).
    id                TEXT        NOT NULL PRIMARY KEY,

    -- Optional username, separate from Kratos traits so it can be changed
    -- without touching the identity provider.
    username          TEXT        UNIQUE,

    -- URL to the stored avatar image (local path or S3 URL).
    -- NULL means no avatar has been uploaded yet.
    avatar_url        TEXT,

    -- Free-form biography displayed on the user's profile.
    bio               TEXT,

    -- UI theme preference. Validated at the application layer.
    -- Values: 'light' | 'dark' | 'system'
    theme_preference  TEXT        NOT NULL DEFAULT 'system',

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Auto-update updated_at on every row change.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_profiles_set_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Index to support future queries like "list users by creation date".
CREATE INDEX IF NOT EXISTS idx_user_profiles_created_at ON user_profiles (created_at DESC);
