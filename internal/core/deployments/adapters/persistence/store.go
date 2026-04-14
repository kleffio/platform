package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/deployments/domain"
	"github.com/kleffio/platform/internal/core/deployments/ports"
)

// PostgresDeploymentStore implements ports.DeploymentRepository against PostgreSQL.
type PostgresDeploymentStore struct {
	db *sql.DB
}

func NewPostgresDeploymentStore(db *sql.DB) ports.DeploymentRepository {
	return &PostgresDeploymentStore{db: db}
}

func (s *PostgresDeploymentStore) Save(ctx context.Context, d *domain.Deployment) error {
	var finishedAt *time.Time
	if d.FinishedAt != nil {
		finishedAt = d.FinishedAt
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO deployments
			(id, organization_id, game_server_id, server_name, blueprint_id, version, status, initiated_by, failure_reason, address, started_at, finished_at, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET
			status         = EXCLUDED.status,
			server_name    = EXCLUDED.server_name,
			failure_reason = EXCLUDED.failure_reason,
			address        = EXCLUDED.address,
			finished_at    = EXCLUDED.finished_at,
			updated_at     = EXCLUDED.updated_at`,
		d.ID, d.OrganizationID, d.GameServerID, d.ServerName, d.BlueprintID, d.Version,
		string(d.Status), d.InitiatedBy, d.FailureReason, d.Address,
		d.StartedAt, finishedAt, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save deployment: %w", err)
	}
	return nil
}

func (s *PostgresDeploymentStore) FindByID(ctx context.Context, id string) (*domain.Deployment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, game_server_id, server_name, blueprint_id, version, status, initiated_by, failure_reason, address, started_at, finished_at, created_at, updated_at
		FROM deployments WHERE id = $1`, id)
	d, err := scanDeployment(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment not found: %s", id)
	}
	return d, err
}

func (s *PostgresDeploymentStore) FindByServerID(ctx context.Context, serverID string) (*domain.Deployment, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, game_server_id, server_name, blueprint_id, version, status, initiated_by, failure_reason, address, started_at, finished_at, created_at, updated_at
		FROM deployments WHERE game_server_id = $1
		ORDER BY created_at DESC LIMIT 1`, serverID)
	d, err := scanDeployment(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment not found for server: %s", serverID)
	}
	return d, err
}

func (s *PostgresDeploymentStore) UpdateAddress(ctx context.Context, serverID, address string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		UPDATE deployments
		SET address = $1, status = $2, finished_at = $3, updated_at = $3
		WHERE game_server_id = $4`,
		address, string(domain.DeploymentSucceeded), now, serverID,
	)
	if err != nil {
		return fmt.Errorf("update address: %w", err)
	}
	return nil
}

func (s *PostgresDeploymentStore) UpdateStatus(ctx context.Context, serverID, status string) error {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx, `
		UPDATE deployments
		SET status = $1, updated_at = $2
		WHERE game_server_id = $3`,
		status, now, serverID,
	)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	return nil
}

func (s *PostgresDeploymentStore) ListByGameServer(ctx context.Context, gameServerID string, page, limit int) ([]*domain.Deployment, int, error) {
	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, game_server_id, server_name, blueprint_id, version, status, initiated_by, failure_reason, address, started_at, finished_at, created_at, updated_at
		FROM deployments WHERE game_server_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, gameServerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list deployments: %w", err)
	}
	defer rows.Close()

	var out []*domain.Deployment
	for rows.Next() {
		d, err := scanDeployment(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}

	var total int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployments WHERE game_server_id = $1`, gameServerID).Scan(&total)
	return out, total, rows.Err()
}

func (s *PostgresDeploymentStore) ListByOrganization(ctx context.Context, orgID string, page, limit int) ([]*domain.Deployment, int, error) {
	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, game_server_id, server_name, blueprint_id, version, status, initiated_by, failure_reason, address, started_at, finished_at, created_at, updated_at
		FROM deployments WHERE organization_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, orgID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list deployments: %w", err)
	}
	defer rows.Close()

	var out []*domain.Deployment
	for rows.Next() {
		d, err := scanDeployment(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}

	var total int
	_ = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM deployments WHERE organization_id = $1`, orgID).Scan(&total)
	return out, total, rows.Err()
}

func (s *PostgresDeploymentStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM deployments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete deployment: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanDeployment(s scanner) (*domain.Deployment, error) {
	var d domain.Deployment
	var status string
	var finishedAt *time.Time
	err := s.Scan(
		&d.ID, &d.OrganizationID, &d.GameServerID, &d.ServerName, &d.BlueprintID, &d.Version,
		&status, &d.InitiatedBy, &d.FailureReason, &d.Address,
		&d.StartedAt, &finishedAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.Status = domain.DeploymentStatus(status)
	d.FinishedAt = finishedAt
	return &d, nil
}
