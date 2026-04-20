package persistence

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/organizations/domain"
	"github.com/kleffio/platform/internal/core/organizations/ports"
)

type PostgresOrgStore struct {
	db *sql.DB
}

func NewPostgresOrgStore(db *sql.DB) ports.OrganizationRepository {
	return &PostgresOrgStore{db: db}
}

// ── Organizations ─────────────────────────────────────────────────────────────

func (s *PostgresOrgStore) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(slug, ''), created_at, updated_at
		FROM organizations WHERE id = $1`, id)
	return scanOrg(row)
}

func (s *PostgresOrgStore) FindBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(slug, ''), created_at, updated_at
		FROM organizations WHERE slug = $1`, slug)
	return scanOrg(row)
}

func (s *PostgresOrgStore) Save(ctx context.Context, org *domain.Organization) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)`,
		org.ID, org.Name, org.CreatedAt, org.UpdatedAt)
	if err != nil {
		return fmt.Errorf("save organization: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) Update(ctx context.Context, org *domain.Organization) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE organizations SET name = $1, updated_at = $2 WHERE id = $3`,
		org.Name, org.UpdatedAt, org.ID)
	if err != nil {
		return fmt.Errorf("update organization: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete organization: %w", err)
	}
	return nil
}

// ── Membership ────────────────────────────────────────────────────────────────

func (s *PostgresOrgStore) ListByUserID(ctx context.Context, userID string) ([]*domain.Organization, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT o.id, o.name, COALESCE(o.slug, ''), o.created_at, o.updated_at
		FROM organizations o
		JOIN organization_members m ON m.org_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.created_at ASC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list orgs by user: %w", err)
	}
	defer rows.Close()
	return scanOrgs(rows)
}

func (s *PostgresOrgStore) ListMembers(ctx context.Context, orgID string) ([]*domain.Member, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT org_id, user_id, email, display_name, role, created_at
		FROM organization_members
		WHERE org_id = $1
		ORDER BY created_at ASC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()
	return scanMembers(rows)
}

func (s *PostgresOrgStore) GetMember(ctx context.Context, orgID, userID string) (*domain.Member, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT org_id, user_id, email, display_name, role, created_at
		FROM organization_members
		WHERE org_id = $1 AND user_id = $2`, orgID, userID)
	return scanMember(row)
}

func (s *PostgresOrgStore) AddMember(ctx context.Context, m *domain.Member) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO organization_members (org_id, user_id, email, display_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (org_id, user_id) DO NOTHING`,
		m.OrgID, m.UserID, m.Email, m.DisplayName, m.Role, m.CreatedAt)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE organization_members SET role = $1
		WHERE org_id = $2 AND user_id = $3`,
		role, orgID, userID)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) RemoveMember(ctx context.Context, orgID, userID string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2`,
		orgID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) CountOwners(ctx context.Context, orgID string) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM organization_members
		WHERE org_id = $1 AND role = 'owner'`, orgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count owners: %w", err)
	}
	return count, nil
}

// EnsureOrgWithOwner upserts the organization row and adds the caller as owner
// if they are not already a member. Handles the personal-org bootstrap path.
func (s *PostgresOrgStore) EnsureOrgWithOwner(ctx context.Context, orgID, orgName, userID, email, displayName string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	now := time.Now().UTC()

	// Upsert the org row.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO organizations (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET updated_at = EXCLUDED.updated_at`,
		orgID, orgName, now, now)
	if err != nil {
		return fmt.Errorf("upsert organization: %w", err)
	}

	// Add the caller as owner only if they have no membership row yet.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_members (org_id, user_id, email, display_name, role, created_at)
		VALUES ($1, $2, $3, $4, 'owner', $5)
		ON CONFLICT (org_id, user_id) DO NOTHING`,
		orgID, userID, email, displayName, now)
	if err != nil {
		return fmt.Errorf("ensure owner membership: %w", err)
	}

	return tx.Commit()
}

// ── Invites ───────────────────────────────────────────────────────────────────

func (s *PostgresOrgStore) CreateInvite(ctx context.Context, inv *domain.Invite) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO org_invites (id, org_id, invited_email, role, token_hash, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		inv.ID, inv.OrgID, inv.InvitedEmail, inv.Role, inv.TokenHash,
		inv.InvitedBy, inv.ExpiresAt, inv.CreatedAt)
	if err != nil {
		return fmt.Errorf("create invite: %w", err)
	}
	return nil
}

func (s *PostgresOrgStore) FindInviteByToken(ctx context.Context, tokenHash string) (*domain.Invite, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, org_id, invited_email, role, token_hash, invited_by,
		       expires_at, accepted_at, created_at
		FROM org_invites WHERE token_hash = $1`, tokenHash)
	return scanInvite(row)
}

func (s *PostgresOrgStore) ListInvites(ctx context.Context, orgID string) ([]*domain.Invite, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, org_id, invited_email, role, token_hash, invited_by,
		       expires_at, accepted_at, created_at
		FROM org_invites
		WHERE org_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	defer rows.Close()
	return scanInvites(rows)
}

// AcceptInvite marks the invite as accepted and adds the user as a member.
func (s *PostgresOrgStore) AcceptInvite(ctx context.Context, inviteID, userID, email, displayName string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var inv struct {
		OrgID string
		Role  string
	}
	err = tx.QueryRowContext(ctx, `
		SELECT org_id, role FROM org_invites
		WHERE id = $1 AND accepted_at IS NULL AND expires_at > NOW()`, inviteID).
		Scan(&inv.OrgID, &inv.Role)
	if err == sql.ErrNoRows {
		return fmt.Errorf("invite not found or already used")
	}
	if err != nil {
		return fmt.Errorf("find invite: %w", err)
	}

	now := time.Now().UTC()

	_, err = tx.ExecContext(ctx, `
		UPDATE org_invites SET accepted_at = $1 WHERE id = $2`, now, inviteID)
	if err != nil {
		return fmt.Errorf("mark invite accepted: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_members (org_id, user_id, email, display_name, role, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		inv.OrgID, userID, email, displayName, inv.Role, now)
	if err != nil {
		return fmt.Errorf("add member from invite: %w", err)
	}

	return tx.Commit()
}

func (s *PostgresOrgStore) RevokeInvite(ctx context.Context, inviteID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM org_invites WHERE id = $1`, inviteID)
	if err != nil {
		return fmt.Errorf("revoke invite: %w", err)
	}
	return nil
}

// ── Scanners ──────────────────────────────────────────────────────────────────

func scanOrg(row *sql.Row) (*domain.Organization, error) {
	var o domain.Organization
	if err := row.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func scanOrgs(rows *sql.Rows) ([]*domain.Organization, error) {
	var out []*domain.Organization
	for rows.Next() {
		var o domain.Organization
		if err := rows.Scan(&o.ID, &o.Name, &o.Slug, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &o)
	}
	return out, rows.Err()
}

func scanMember(row *sql.Row) (*domain.Member, error) {
	var m domain.Member
	if err := row.Scan(&m.OrgID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
		return nil, err
	}
	return &m, nil
}

func scanMembers(rows *sql.Rows) ([]*domain.Member, error) {
	var out []*domain.Member
	for rows.Next() {
		var m domain.Member
		if err := rows.Scan(&m.OrgID, &m.UserID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func scanInvite(row *sql.Row) (*domain.Invite, error) {
	var inv domain.Invite
	var acceptedAt sql.NullTime
	if err := row.Scan(&inv.ID, &inv.OrgID, &inv.InvitedEmail, &inv.Role,
		&inv.TokenHash, &inv.InvitedBy, &inv.ExpiresAt, &acceptedAt, &inv.CreatedAt); err != nil {
		return nil, err
	}
	if acceptedAt.Valid {
		inv.AcceptedAt = &acceptedAt.Time
	}
	return &inv, nil
}

func scanInvites(rows *sql.Rows) ([]*domain.Invite, error) {
	var out []*domain.Invite
	for rows.Next() {
		var inv domain.Invite
		var acceptedAt sql.NullTime
		if err := rows.Scan(&inv.ID, &inv.OrgID, &inv.InvitedEmail, &inv.Role,
			&inv.TokenHash, &inv.InvitedBy, &inv.ExpiresAt, &acceptedAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		if acceptedAt.Valid {
			inv.AcceptedAt = &acceptedAt.Time
		}
		out = append(out, &inv)
	}
	return out, rows.Err()
}

// HashToken returns the SHA-256 hex of a raw token for safe DB storage.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
