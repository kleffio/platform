-- 006_projects_collaboration.sql
-- Adds workload connections and canvas graph positions for project architecture.

-- project_connections
-- Represents a logical connection between two workloads in the architecture canvas.
-- kind: 'network' | 'dependency' | 'traffic'
CREATE TABLE IF NOT EXISTS project_connections (
    id                  TEXT PRIMARY KEY,
    project_id          TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source_workload_id  TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    target_workload_id  TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    kind                TEXT NOT NULL DEFAULT 'network',
    label               TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_connections_kind_check
        CHECK (kind IN ('network', 'dependency', 'traffic')),
    CONSTRAINT project_connections_unique
        UNIQUE (project_id, source_workload_id, target_workload_id, kind)
);

CREATE INDEX IF NOT EXISTS idx_project_connections_project_id
    ON project_connections(project_id);

-- project_graph_nodes
-- Persists canvas (x,y) positions for each workload node so the layout survives
-- page reloads.
CREATE TABLE IF NOT EXISTS project_graph_nodes (
    id          TEXT PRIMARY KEY,
    project_id  TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    workload_id TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    position_x  DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y  DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT project_graph_nodes_unique UNIQUE (project_id, workload_id)
);

CREATE INDEX IF NOT EXISTS idx_project_graph_nodes_project_id
    ON project_graph_nodes(project_id);
