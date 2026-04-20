CREATE TABLE IF NOT EXISTS usage_records (
    id               TEXT        PRIMARY KEY,
    organization_id  TEXT        NOT NULL,
    workload_id      TEXT        NOT NULL,
    node_id          TEXT        NOT NULL,
    recorded_at      TIMESTAMPTZ NOT NULL,
    cpu_seconds      DOUBLE PRECISION NOT NULL DEFAULT 0,
    memory_gb_hours  DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_in_mb    DOUBLE PRECISION NOT NULL DEFAULT 0,
    network_out_mb   DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_read_mb     DOUBLE PRECISION NOT NULL DEFAULT 0,
    disk_write_mb    DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_usage_records_workload ON usage_records (workload_id, recorded_at DESC);
CREATE INDEX IF NOT EXISTS idx_usage_records_org ON usage_records (organization_id, recorded_at DESC);
