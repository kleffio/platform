-- 010_project_members_invites.sql
-- Enhances project_members with profile columns and adds project_invites table.

ALTER TABLE project_members
    ADD COLUMN IF NOT EXISTS email        TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS display_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS invited_by   TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_project_members_user_id ON project_members(user_id);

-- Pending email invitations into a project.
CREATE TABLE IF NOT EXISTS project_invites (
    id            TEXT        PRIMARY KEY,
    project_id    TEXT        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    invited_email TEXT        NOT NULL,
    role          TEXT        NOT NULL DEFAULT 'developer',
    token_hash    TEXT        NOT NULL UNIQUE,
    invited_by    TEXT        NOT NULL DEFAULT '',
    expires_at    TIMESTAMPTZ NOT NULL,
    accepted_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_invites_role_check
        CHECK (role IN ('owner', 'maintainer', 'developer', 'viewer'))
);

CREATE INDEX IF NOT EXISTS idx_project_invites_project_id ON project_invites(project_id);
CREATE INDEX IF NOT EXISTS idx_project_invites_token      ON project_invites(token_hash);
