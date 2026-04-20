package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kleffio/platform/internal/core/usage/domain"
)

type PostgresUsageStore struct {
	db *sql.DB
}

func NewPostgresUsageStore(db *sql.DB) *PostgresUsageStore {
	return &PostgresUsageStore{db: db}
}

func (s *PostgresUsageStore) Save(ctx context.Context, r *domain.UsageRecord) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO usage_records (
			id, organization_id, project_id, workload_id, node_id, recorded_at,
			cpu_seconds, memory_gb_hours, network_in_mb, network_out_mb,
			disk_read_mb, disk_write_mb,
			cpu_millicores, memory_mb,
			network_in_kbps, network_out_kbps,
			disk_read_kbps, disk_write_kbps
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
	`,
		r.ID,
		r.OrganizationID,
		r.ProjectID,
		r.GameServerID,
		r.NodeID,
		r.RecordedAt,
		r.CPUSeconds,
		r.MemoryGBHours,
		r.NetworkInMB,
		r.NetworkOutMB,
		r.DiskReadMB,
		r.DiskWriteMB,
		r.CPUMillicores,
		r.MemoryMB,
		r.NetworkInKbps,
		r.NetworkOutKbps,
		r.DiskReadKbps,
		r.DiskWriteKbps,
	)
	if err != nil {
		return fmt.Errorf("save usage record: %w", err)
	}
	return nil
}

// ListLatestByProject returns the most recent metrics snapshot per workload for a given project,
// joined with the workload's allocated CPU/memory limits.
func (s *PostgresUsageStore) ListLatestByProject(ctx context.Context, projectID string) ([]*domain.WorkloadMetrics, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT
			u.workload_id, u.project_id,
			u.cpu_millicores, u.memory_mb,
			u.network_in_kbps, u.network_out_kbps,
			u.disk_read_kbps, u.disk_write_kbps,
			u.recorded_at,
			COALESCE(w.cpu_millicores, 0),
			COALESCE(w.memory_bytes, 0)
		FROM (
			SELECT DISTINCT ON (workload_id)
				workload_id, project_id,
				cpu_millicores, memory_mb,
				network_in_kbps, network_out_kbps,
				disk_read_kbps, disk_write_kbps,
				recorded_at
			FROM usage_records
			WHERE project_id = $1
			ORDER BY workload_id, recorded_at DESC
		) u
		JOIN workloads w ON w.id = u.workload_id AND w.state != 'deleted'
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list latest usage by project: %w", err)
	}
	defer rows.Close()

	var results []*domain.WorkloadMetrics
	for rows.Next() {
		m := &domain.WorkloadMetrics{}
		if err := rows.Scan(
			&m.WorkloadID, &m.ProjectID,
			&m.CPUMillicores, &m.MemoryMB,
			&m.NetworkInKbps, &m.NetworkOutKbps,
			&m.DiskReadKbps, &m.DiskWriteKbps,
			&m.RecordedAt,
			&m.CPULimitMillicores, &m.MemoryLimitBytes,
		); err != nil {
			return nil, fmt.Errorf("scan usage row: %w", err)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}
