ALTER TABLE usage_records
    ADD COLUMN IF NOT EXISTS project_id        TEXT             NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cpu_millicores    BIGINT           NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS memory_mb         BIGINT           NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS network_in_kbps   DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS network_out_kbps  DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS disk_read_kbps    DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS disk_write_kbps   DOUBLE PRECISION NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_usage_records_project_recorded
    ON usage_records (project_id, recorded_at DESC);
