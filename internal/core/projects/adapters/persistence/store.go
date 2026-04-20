package persistence

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/projects/domain"
	"github.com/kleffio/platform/internal/core/projects/ports"
	"github.com/kleffio/platform/internal/shared/ids"
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

func (s *PostgresProjectStore) ListByMember(ctx context.Context, userID string) ([]*domain.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT p.id, p.organization_id, p.slug, p.name, p.is_default, p.created_at, p.updated_at
		FROM projects p
		INNER JOIN project_members pm ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY p.created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list projects by member: %w", err)
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

// ── Project members ───────────────────────────────────────────────────────────

func (s *PostgresProjectStore) ListMembers(ctx context.Context, projectID string) ([]*domain.ProjectMember, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT project_id, user_id, email, display_name, role, invited_by, created_at
		FROM project_members WHERE project_id = $1 ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project members: %w", err)
	}
	defer rows.Close()
	var out []*domain.ProjectMember
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *PostgresProjectStore) GetMember(ctx context.Context, projectID, userID string) (*domain.ProjectMember, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT project_id, user_id, email, display_name, role, invited_by, created_at
		FROM project_members WHERE project_id = $1 AND user_id = $2`, projectID, userID)
	return scanMember(row)
}

func (s *PostgresProjectStore) AddMember(ctx context.Context, m *domain.ProjectMember) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, email, display_name, role, invited_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (project_id, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			display_name = EXCLUDED.display_name,
			email = EXCLUDED.email`,
		m.ProjectID, m.UserID, m.Email, m.DisplayName, m.Role, m.InvitedBy, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("add project member: %w", err)
	}
	return nil
}

func (s *PostgresProjectStore) UpdateMemberRole(ctx context.Context, projectID, userID, role string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE project_members SET role=$3 WHERE project_id=$1 AND user_id=$2`, projectID, userID, role)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *PostgresProjectStore) RemoveMember(ctx context.Context, projectID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM project_members WHERE project_id=$1 AND user_id=$2`, projectID, userID)
	return err
}

// ── Project invites ───────────────────────────────────────────────────────────

func (s *PostgresProjectStore) ListInvites(ctx context.Context, projectID string) ([]*domain.ProjectInvite, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, project_id, invited_email, role, invited_by, expires_at, accepted_at, created_at
		FROM project_invites
		WHERE project_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project invites: %w", err)
	}
	defer rows.Close()
	var out []*domain.ProjectInvite
	for rows.Next() {
		inv, err := scanInvite(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (s *PostgresProjectStore) FindInviteByToken(ctx context.Context, tokenHash string) (*domain.ProjectInvite, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, invited_email, role, invited_by, expires_at, accepted_at, created_at
		FROM project_invites WHERE token_hash = $1`, tokenHash)
	return scanInvite(row)
}

func (s *PostgresProjectStore) FindActiveInviteByEmail(ctx context.Context, projectID, email string) (*domain.ProjectInvite, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, project_id, invited_email, role, invited_by, expires_at, accepted_at, created_at
		FROM project_invites
		WHERE project_id = $1 AND LOWER(invited_email) = LOWER($2)
		  AND accepted_at IS NULL AND expires_at > NOW()
		LIMIT 1`, projectID, email)
	return scanInvite(row)
}

func (s *PostgresProjectStore) CreateInvite(ctx context.Context, inv *domain.ProjectInvite) error {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return fmt.Errorf("generate invite token: %w", err)
	}
	token := hex.EncodeToString(raw)
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])

	inv.ID = ids.New()
	inv.Token = token
	inv.TokenHash = tokenHash
	if inv.CreatedAt.IsZero() {
		inv.CreatedAt = time.Now().UTC()
	}
	if inv.ExpiresAt.IsZero() {
		inv.ExpiresAt = inv.CreatedAt.Add(7 * 24 * time.Hour)
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO project_invites (id, project_id, invited_email, role, token_hash, invited_by, expires_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		inv.ID, inv.ProjectID, inv.InvitedEmail, inv.Role, tokenHash, inv.InvitedBy, inv.ExpiresAt, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("create project invite: %w", err)
	}
	return nil
}

func (s *PostgresProjectStore) AcceptInvite(ctx context.Context, tokenHash, userID, email, displayName string) (*domain.ProjectInvite, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var inv domain.ProjectInvite
	var acceptedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		SELECT id, project_id, invited_email, role, invited_by, expires_at, accepted_at, created_at
		FROM project_invites WHERE token_hash = $1 FOR UPDATE`, tokenHash).Scan(
		&inv.ID, &inv.ProjectID, &inv.InvitedEmail, &inv.Role, &inv.InvitedBy, &inv.ExpiresAt, &acceptedAt, &inv.CreatedAt)
	if err != nil {
		return nil, err
	}
	if acceptedAt.Valid {
		return nil, fmt.Errorf("invite already accepted")
	}
	if time.Now().After(inv.ExpiresAt) {
		return nil, fmt.Errorf("invite expired")
	}

	now := time.Now().UTC()
	if _, err := tx.ExecContext(ctx,
		`UPDATE project_invites SET accepted_at=$2 WHERE id=$1`, inv.ID, now); err != nil {
		return nil, fmt.Errorf("mark invite accepted: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO project_members (project_id, user_id, email, display_name, role, invited_by, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (project_id, user_id) DO UPDATE SET role=EXCLUDED.role`,
		inv.ProjectID, userID, email, displayName, inv.Role, inv.InvitedBy, now); err != nil {
		return nil, fmt.Errorf("add member on accept: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit accept invite: %w", err)
	}

	inv.AcceptedAt = &now
	return &inv, nil
}

func (s *PostgresProjectStore) RevokeInvite(ctx context.Context, projectID, inviteID string) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM project_invites WHERE id=$1 AND project_id=$2 AND accepted_at IS NULL`, inviteID, projectID)
	if err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
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

func scanMember(s scanner) (*domain.ProjectMember, error) {
	var m domain.ProjectMember
	if err := s.Scan(&m.ProjectID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.InvitedBy, &m.CreatedAt); err != nil {
		return nil, err
	}
	m.CreatedAt = m.CreatedAt.UTC()
	return &m, nil
}

func scanInvite(s scanner) (*domain.ProjectInvite, error) {
	var inv domain.ProjectInvite
	var acceptedAt sql.NullTime
	if err := s.Scan(&inv.ID, &inv.ProjectID, &inv.InvitedEmail, &inv.Role, &inv.InvitedBy, &inv.ExpiresAt, &acceptedAt, &inv.CreatedAt); err != nil {
		return nil, err
	}
	inv.ExpiresAt = inv.ExpiresAt.UTC()
	inv.CreatedAt = inv.CreatedAt.UTC()
	if acceptedAt.Valid {
		t := acceptedAt.Time.UTC()
		inv.AcceptedAt = &t
	}
	return &inv, nil
}
