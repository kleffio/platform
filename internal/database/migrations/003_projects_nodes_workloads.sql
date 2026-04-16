CREATE TABLE IF NOT EXISTS organizations (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    slug            TEXT NOT NULL,
    name            TEXT NOT NULL,
    is_default      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, slug)
);

CREATE TABLE IF NOT EXISTS project_members (
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id     TEXT NOT NULL,
    role        TEXT NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS nodes (
    id                TEXT PRIMARY KEY,
    hostname          TEXT NOT NULL UNIQUE,
    region            TEXT NOT NULL DEFAULT 'local',
    ip_address        TEXT NOT NULL DEFAULT '',
    status            TEXT NOT NULL DEFAULT 'online',
    total_vcpu        INTEGER NOT NULL DEFAULT 0,
    total_mem_gb      INTEGER NOT NULL DEFAULT 0,
    total_disk_gb     INTEGER NOT NULL DEFAULT 0,
    used_vcpu         INTEGER NOT NULL DEFAULT 0,
    used_mem_gb       INTEGER NOT NULL DEFAULT 0,
    used_disk_gb      INTEGER NOT NULL DEFAULT 0,
    token_hash        TEXT NOT NULL DEFAULT '',
    last_heartbeat_at TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_nodes_token_hash
    ON nodes(token_hash)
    WHERE token_hash <> '';

CREATE TABLE IF NOT EXISTS workloads (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id      TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    owner_id        TEXT NOT NULL DEFAULT '',
    blueprint_id    TEXT NOT NULL DEFAULT '',
    image           TEXT NOT NULL DEFAULT '',
    runtime_ref     TEXT NOT NULL DEFAULT '',
    endpoint        TEXT NOT NULL DEFAULT '',
    node_id         TEXT REFERENCES nodes(id) ON DELETE SET NULL,
    state           TEXT NOT NULL DEFAULT 'pending',
    error_message   TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_workloads_project_id ON workloads(project_id);
CREATE INDEX IF NOT EXISTS idx_workloads_state ON workloads(state);

CREATE TABLE IF NOT EXISTS deployments (
    id              TEXT PRIMARY KEY,
    organization_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    project_id      TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    workload_id     TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    game_server_id  TEXT NOT NULL DEFAULT '',
    version         TEXT NOT NULL DEFAULT '',
    action          TEXT NOT NULL DEFAULT 'provision',
    status          TEXT NOT NULL DEFAULT 'pending',
    initiated_by    TEXT NOT NULL DEFAULT '',
    failure_reason  TEXT NOT NULL DEFAULT '',
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE deployments ADD COLUMN IF NOT EXISTS project_id  TEXT NOT NULL DEFAULT '';
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS workload_id TEXT NOT NULL DEFAULT '';
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS action      TEXT NOT NULL DEFAULT 'provision';

CREATE INDEX IF NOT EXISTS idx_deployments_workload_id ON deployments(workload_id);
CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
