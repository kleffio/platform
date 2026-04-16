-- Add a user-visible name to workloads so the globally-unique workload ID can
-- diverge from the human-readable server name, enabling the same name to be
-- reused across different projects.

ALTER TABLE workloads ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

-- Back-fill: for existing rows the ID *was* the server name.
UPDATE workloads SET name = id WHERE name = '';

-- Enforce per-project uniqueness on the name (not the ID).
CREATE UNIQUE INDEX IF NOT EXISTS workloads_project_name_unique
    ON workloads (project_id, name)
    WHERE state <> 'deleted';
