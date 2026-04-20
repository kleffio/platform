-- Stores raw log lines shipped from daemon containers.
CREATE TABLE IF NOT EXISTS workload_log_lines (
    id          BIGSERIAL PRIMARY KEY,
    workload_id TEXT        NOT NULL,
    project_id  TEXT        NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    stream      TEXT        NOT NULL DEFAULT 'stdout', -- 'stdout' or 'stderr'
    line        TEXT        NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_wll_workload_ts ON workload_log_lines (workload_id, ts DESC);
CREATE INDEX IF NOT EXISTS idx_wll_project_ts  ON workload_log_lines (project_id, ts DESC);
