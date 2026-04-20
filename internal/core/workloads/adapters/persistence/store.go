package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/workloads/domain"
	"github.com/kleffio/platform/internal/core/workloads/ports"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) ports.Repository {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) CreateWorkload(ctx context.Context, workload *domain.Workload) error {
	if workload == nil {
		return fmt.Errorf("workload is required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO workloads (
			id, name, organization_id, project_id, owner_id, blueprint_id,
			image, runtime_ref, endpoint, node_id, state, error_message,
			cpu_millicores, memory_bytes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11, $12,
			$13, $14,
			$15, $16
		)`,
		workload.ID,
		workload.Name,
		workload.OrganizationID,
		workload.ProjectID,
		workload.OwnerID,
		workload.BlueprintID,
		workload.Image,
		workload.RuntimeRef,
		workload.Endpoint,
		nullIfEmpty(workload.NodeID),
		workload.State,
		workload.ErrorMessage,
		workload.CPUMillicores,
		workload.MemoryBytes,
		workload.CreatedAt,
		workload.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create workload: %w", err)
	}
	return nil
}

func (s *PostgresStore) FindByProjectAndName(ctx context.Context, projectID, name string) (*domain.Workload, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, organization_id, project_id, owner_id, blueprint_id,
		       image, runtime_ref, endpoint, COALESCE(node_id, ''), state,
		       error_message, cpu_millicores, memory_bytes, created_at, updated_at
		FROM workloads
		WHERE project_id = $1 AND name = $2
		ORDER BY updated_at DESC
		LIMIT 1`, projectID, name)
	return scanWorkload(row)
}

func (s *PostgresStore) FindByID(ctx context.Context, workloadID string) (*domain.Workload, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, organization_id, project_id, owner_id, blueprint_id,
		       image, runtime_ref, endpoint, COALESCE(node_id, ''), state,
		       error_message, cpu_millicores, memory_bytes, created_at, updated_at
		FROM workloads WHERE id = $1`, workloadID)
	return scanWorkload(row)
}

func (s *PostgresStore) ListByProject(ctx context.Context, projectID string) ([]*domain.Workload, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, organization_id, project_id, owner_id, blueprint_id,
		       image, runtime_ref, endpoint, COALESCE(node_id, ''), state,
		       error_message, cpu_millicores, memory_bytes, created_at, updated_at
		FROM workloads
		WHERE project_id = $1
		ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list workloads: %w", err)
	}
	defer rows.Close()

	var out []*domain.Workload
	for rows.Next() {
		workload, err := scanWorkload(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, workload)
	}
	return out, rows.Err()
}

func (s *PostgresStore) SaveDeployment(ctx context.Context, d *ports.DeploymentRecord) error {
	if d == nil {
		return fmt.Errorf("deployment is required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO deployments (
			id, organization_id, project_id, workload_id,
			game_server_id, version, action, status,
			initiated_by, failure_reason,
			started_at, finished_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			'', '', $5, $6,
			$7, '',
			NOW(), NULL, NOW(), NOW()
		)`,
		d.ID,
		d.OrganizationID,
		d.ProjectID,
		d.WorkloadID,
		d.Action,
		d.Status,
		d.InitiatedBy,
	)
	if err != nil {
		return fmt.Errorf("save deployment: %w", err)
	}
	return nil
}

func (s *PostgresStore) DeleteWorkload(ctx context.Context, workloadID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("delete workload: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM deployments WHERE workload_id = $1`, workloadID); err != nil {
		return fmt.Errorf("delete workload deployments: %w", err)
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM workloads WHERE id = $1`, workloadID)
	if err != nil {
		return fmt.Errorf("delete workload: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete workload: rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("delete workload: commit: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateState(ctx context.Context, workloadID string, state domain.WorkloadState, errorMessage string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE workloads
		SET state = $2,
		    error_message = $3,
		    updated_at = NOW()
		WHERE id = $1`,
		workloadID,
		state,
		errorMessage,
	)
	if err != nil {
		return fmt.Errorf("update workload state: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update workload state: rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresStore) UpdateFromDaemon(ctx context.Context, update domain.DaemonStatusUpdate) error {
	observedAt := update.ObservedAt
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE workloads
		SET state = $2,
		    runtime_ref = CASE WHEN $3 = '' THEN runtime_ref ELSE $3 END,
		    endpoint = CASE WHEN $4 = '' THEN endpoint ELSE $4 END,
		    node_id = CASE WHEN $5 = '' THEN node_id ELSE $5 END,
		    error_message = $6,
		    updated_at = $7
		WHERE id = $1`,
		update.WorkloadID,
		update.Status,
		update.RuntimeRef,
		update.Endpoint,
		nullIfEmpty(update.NodeID),
		update.ErrorMessage,
		observedAt,
	)
	if err != nil {
		return fmt.Errorf("update workload from daemon: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update workload from daemon: rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	status := string(update.Status)
	markFinished := status == "running" || status == "stopped" || status == "deleted" || status == "failed"

	_, err = s.db.ExecContext(ctx, `
		UPDATE deployments
		SET status = $2,
		    failure_reason = $3,
		    finished_at = CASE WHEN $4 THEN NOW() ELSE finished_at END,
		    updated_at = NOW()
		WHERE id = (
			SELECT id
			FROM deployments
			WHERE workload_id = $1
			ORDER BY created_at DESC
			LIMIT 1
		)`,
		update.WorkloadID,
		status,
		update.ErrorMessage,
		markFinished,
	)
	if err != nil {
		return fmt.Errorf("update deployment from daemon: %w", err)
	}

	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWorkload(s scanner) (*domain.Workload, error) {
	var w domain.Workload
	if err := s.Scan(
		&w.ID,
		&w.Name,
		&w.OrganizationID,
		&w.ProjectID,
		&w.OwnerID,
		&w.BlueprintID,
		&w.Image,
		&w.RuntimeRef,
		&w.Endpoint,
		&w.NodeID,
		&w.State,
		&w.ErrorMessage,
		&w.CPUMillicores,
		&w.MemoryBytes,
		&w.CreatedAt,
		&w.UpdatedAt,
	); err != nil {
		return nil, err
	}
	w.CreatedAt = w.CreatedAt.UTC()
	w.UpdatedAt = w.UpdatedAt.UTC()
	return &w, nil
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
