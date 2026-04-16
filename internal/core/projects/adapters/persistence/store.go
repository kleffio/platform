package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/projects/domain"
	"github.com/kleffio/platform/internal/core/projects/ports"
)

type PostgresProjectStore struct {
	db *sql.DB
}

func NewPostgresProjectStore(db *sql.DB) ports.ProjectRepository {
	return &PostgresProjectStore{db: db}
}

// ── Organization ─────────────────────────────────────────────────────────────

func (s *PostgresProjectStore) EnsureOrganization(ctx context.Context, organizationID, name string) error {
	if organizationID == "" {
		return fmt.Errorf("organization id is required")
	}
	if name == "" {
		name = "Organization " + organizationID
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET updated_at = NOW()`,
		organizationID,
		name,
	)
	if err != nil {
		return fmt.Errorf("ensure organization: %w", err)
	}
	return nil
}

// ── Projects ──────────────────────────────────────────────────────────────────

func (s *PostgresProjectStore) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, slug, name, is_default, created_at, updated_at
		FROM projects WHERE id = $1`, id)
	return scanProject(row)
}

func (s *PostgresProjectStore) FindBySlug(ctx context.Context, organizationID, slug string) (*domain.Project, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, organization_id, slug, name, is_default, created_at, updated_at
		FROM projects WHERE organization_id = $1 AND slug = $2`, organizationID, slug)
	return scanProject(row)
}

func (s *PostgresProjectStore) ListByOrganization(ctx context.Context, organizationID string) ([]*domain.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, organization_id, slug, name, is_default, created_at, updated_at
		FROM projects
		WHERE organization_id = $1
		ORDER BY created_at ASC`, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var out []*domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *PostgresProjectStore) Save(ctx context.Context, project *domain.Project) error {
	if project == nil {
		return fmt.Errorf("project is required")
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO projects (id, organization_id, slug, name, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			organization_id = EXCLUDED.organization_id,
			slug = EXCLUDED.slug,
			name = EXCLUDED.name,
			is_default = EXCLUDED.is_default,
			updated_at = EXCLUDED.updated_at`,
		project.ID,
		project.OrganizationID,
		project.Slug,
		project.Name,
		project.IsDefault,
		project.CreatedAt,
		project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save project: %w", err)
	}
	return nil
}

// ── Connections ───────────────────────────────────────────────────────────────

func (s *PostgresProjectStore) ListConnections(ctx context.Context, projectID string) ([]*domain.Connection, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, source_workload_id, target_workload_id, kind, label, created_at
		FROM project_connections
		WHERE project_id = $1
		ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list connections: %w", err)
	}
	defer rows.Close()

	var out []*domain.Connection
	for rows.Next() {
		c, err := scanConnection(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *PostgresProjectStore) FindConnection(ctx context.Context, connectionID string) (*domain.Connection, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, source_workload_id, target_workload_id, kind, label, created_at
		FROM project_connections WHERE id = $1`, connectionID)
	return scanConnection(row)
}

func (s *PostgresProjectStore) CreateConnection(ctx context.Context, conn *domain.Connection) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO project_connections
			(id, project_id, source_workload_id, target_workload_id, kind, label, created_at)
		SELECT $1,$2,$3,$4,$5,$6,$7
		WHERE EXISTS (
			SELECT 1 FROM workloads w1
			WHERE w1.id = $3 AND w1.project_id = $2
		)
		  AND EXISTS (
			SELECT 1 FROM workloads w2
			WHERE w2.id = $4 AND w2.project_id = $2
		)`,
		conn.ID,
		conn.ProjectID,
		conn.SourceWorkloadID,
		conn.TargetWorkloadID,
		conn.Kind,
		conn.Label,
		conn.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create connection: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("create connection rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresProjectStore) DeleteConnection(ctx context.Context, connectionID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM project_connections WHERE id = $1`, connectionID)
	if err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}

// ── Graph nodes ───────────────────────────────────────────────────────────────

func (s *PostgresProjectStore) ListGraphNodes(ctx context.Context, projectID string) ([]*domain.GraphNode, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, workload_id, position_x, position_y, updated_at
		FROM project_graph_nodes
		WHERE project_id = $1`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list graph nodes: %w", err)
	}
	defer rows.Close()

	var out []*domain.GraphNode
	for rows.Next() {
		n := &domain.GraphNode{}
		if err := rows.Scan(&n.ID, &n.ProjectID, &n.WorkloadID, &n.PositionX, &n.PositionY, &n.UpdatedAt); err != nil {
			return nil, err
		}
		n.UpdatedAt = n.UpdatedAt.UTC()
		out = append(out, n)
	}
	return out, rows.Err()
}

func (s *PostgresProjectStore) UpsertGraphNode(ctx context.Context, node *domain.GraphNode) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO project_graph_nodes (id, project_id, workload_id, position_x, position_y, updated_at)
		SELECT $1,$2,$3,$4,$5,$6
		WHERE EXISTS (
			SELECT 1 FROM workloads w
			WHERE w.id = $3 AND w.project_id = $2
		)
		ON CONFLICT ON CONSTRAINT project_graph_nodes_unique DO UPDATE SET
			position_x = EXCLUDED.position_x,
			position_y = EXCLUDED.position_y,
			updated_at = EXCLUDED.updated_at`,
		node.ID,
		node.ProjectID,
		node.WorkloadID,
		node.PositionX,
		node.PositionY,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert graph node: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("upsert graph node rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ── Scanners ──────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(s scanner) (*domain.Project, error) {
	var p domain.Project
	if err := s.Scan(
		&p.ID,
		&p.OrganizationID,
		&p.Slug,
		&p.Name,
		&p.IsDefault,
		&p.CreatedAt,
		&p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	p.CreatedAt = p.CreatedAt.UTC()
	p.UpdatedAt = p.UpdatedAt.UTC()
	return &p, nil
}

func scanConnection(s scanner) (*domain.Connection, error) {
	var c domain.Connection
	if err := s.Scan(&c.ID, &c.ProjectID, &c.SourceWorkloadID, &c.TargetWorkloadID, &c.Kind, &c.Label, &c.CreatedAt); err != nil {
		return nil, err
	}
	c.CreatedAt = c.CreatedAt.UTC()
	return &c, nil
}
