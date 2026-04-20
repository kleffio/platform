-- Tracks which users belong to which organizations and their role.
-- Roles: owner | admin | member
-- Multiple owners are allowed per org.

CREATE TABLE IF NOT EXISTS organization_members (
    org_id       TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id      TEXT        NOT NULL,
    email        TEXT        NOT NULL DEFAULT '',
    display_name TEXT        NOT NULL DEFAULT '',
    role         TEXT        NOT NULL DEFAULT 'member',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_org_members_user_id ON organization_members(user_id);

-- Back-fill: for any organization that was auto-created from a user ID
-- (pattern: "org-<slug>"), we cannot reconstruct the original user_id here,
-- so bootstrap happens at runtime on first login (see EnsureOrganization).

-- Pending email invitations into an organization.
CREATE TABLE IF NOT EXISTS org_invites (
    id            TEXT        PRIMARY KEY,
    org_id        TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    invited_email TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'member',
    token_hash    TEXT        NOT NULL UNIQUE,
    invited_by    TEXT        NOT NULL DEFAULT '',
    expires_at    TIMESTAMPTZ NOT NULL,
    accepted_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_org_invites_org_id ON org_invites(org_id);
CREATE INDEX IF NOT EXISTS idx_org_invites_token  ON org_invites(token_hash);
