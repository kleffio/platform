package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kleff/go-common/domain"
	plugindomain "github.com/kleffio/platform/internal/core/plugins/domain"
	"github.com/kleffio/platform/internal/core/plugins/ports"
)

// PostgresPluginStore implements ports.PluginStore against the plugins and
// settings tables (see migrations/003 and 004).
type PostgresPluginStore struct {
	db *sql.DB
}

// NewPostgresPluginStore creates the store. db may be nil during development;
// all methods gracefully no-op or return empty results.
func NewPostgresPluginStore(db *sql.DB) *PostgresPluginStore {
	return &PostgresPluginStore{db: db}
}

var _ ports.PluginStore = (*PostgresPluginStore)(nil)

// ── Plugins ───────────────────────────────────────────────────────────────────

func (s *PostgresPluginStore) FindByID(ctx context.Context, id string) (*plugindomain.Plugin, error) {
	if s.db == nil {
		return nil, domain.ErrNotFound
	}
	const q = `
		SELECT id, type, display_name, image, version, grpc_addr, frontend_url,
		       config, secrets, enabled, installed_at, updated_at
		FROM plugins WHERE id = $1`
	return scanPlugin(s.db.QueryRowContext(ctx, q, id))
}

func (s *PostgresPluginStore) ListAll(ctx context.Context) ([]*plugindomain.Plugin, error) {
	if s.db == nil {
		return nil, nil
	}
	const q = `
		SELECT id, type, display_name, image, version, grpc_addr, frontend_url,
		       config, secrets, enabled, installed_at, updated_at
		FROM plugins ORDER BY installed_at DESC`
	rows, err := s.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("plugin store: list all: %w", err)
	}
	defer rows.Close()
	return scanPlugins(rows)
}

func (s *PostgresPluginStore) ListByType(ctx context.Context, pluginType string) ([]*plugindomain.Plugin, error) {
	if s.db == nil {
		return nil, nil
	}
	const q = `
		SELECT id, type, display_name, image, version, grpc_addr, frontend_url,
		       config, secrets, enabled, installed_at, updated_at
		FROM plugins WHERE type = $1 ORDER BY installed_at DESC`
	rows, err := s.db.QueryContext(ctx, q, pluginType)
	if err != nil {
		return nil, fmt.Errorf("plugin store: list by type: %w", err)
	}
	defer rows.Close()
	return scanPlugins(rows)
}

func (s *PostgresPluginStore) Save(ctx context.Context, p *plugindomain.Plugin) error {
	if s.db == nil {
		return nil
	}
	const q = `
		INSERT INTO plugins
		    (id, type, display_name, image, version, grpc_addr, frontend_url,
		     config, secrets, enabled, installed_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		ON CONFLICT (id) DO UPDATE SET
		    display_name  = EXCLUDED.display_name,
		    image         = EXCLUDED.image,
		    version       = EXCLUDED.version,
		    grpc_addr     = EXCLUDED.grpc_addr,
		    frontend_url  = EXCLUDED.frontend_url,
		    config        = EXCLUDED.config,
		    secrets       = EXCLUDED.secrets,
		    enabled       = EXCLUDED.enabled,
		    updated_at    = EXCLUDED.updated_at`

	var frontendURL sql.NullString
	if p.FrontendURL != "" {
		frontendURL = sql.NullString{String: p.FrontendURL, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, q,
		p.ID, p.Type, p.DisplayName, p.Image, p.Version, p.GRPCAddr, frontendURL,
		json.RawMessage(p.Config), json.RawMessage(p.Secrets),
		p.Enabled, p.InstalledAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("plugin store: save %q: %w", p.ID, err)
	}
	return nil
}

func (s *PostgresPluginStore) Delete(ctx context.Context, id string) error {
	if s.db == nil {
		return nil
	}
	res, err := s.db.ExecContext(ctx, `DELETE FROM plugins WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("plugin store: delete %q: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ── Settings ──────────────────────────────────────────────────────────────────

func (s *PostgresPluginStore) GetSetting(ctx context.Context, key string) (string, error) {
	if s.db == nil {
		return "", nil
	}
	var val string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM settings WHERE key = $1`, key,
	).Scan(&val)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("plugin store: get setting %q: %w", key, err)
	}
	return val, nil
}

func (s *PostgresPluginStore) SetSetting(ctx context.Context, key, value string) error {
	if s.db == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO settings (key, value) VALUES ($1, $2)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("plugin store: set setting %q: %w", key, err)
	}
	return nil
}

// ── Scan helpers ──────────────────────────────────────────────────────────────

func scanPlugin(row *sql.Row) (*plugindomain.Plugin, error) {
	var (
		p           plugindomain.Plugin
		frontendURL sql.NullString
		config      []byte
		secrets     []byte
		installedAt time.Time
		updatedAt   time.Time
	)
	err := row.Scan(
		&p.ID, &p.Type, &p.DisplayName, &p.Image, &p.Version, &p.GRPCAddr, &frontendURL,
		&config, &secrets, &p.Enabled, &installedAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scan plugin: %w", err)
	}
	p.FrontendURL = frontendURL.String
	p.Config = json.RawMessage(config)
	p.Secrets = json.RawMessage(secrets)
	p.InstalledAt = installedAt.UTC()
	p.UpdatedAt = updatedAt.UTC()
	p.Status = plugindomain.PluginStatusUnknown
	return &p, nil
}

func scanPlugins(rows *sql.Rows) ([]*plugindomain.Plugin, error) {
	var out []*plugindomain.Plugin
	for rows.Next() {
		var (
			p           plugindomain.Plugin
			frontendURL sql.NullString
			config      []byte
			secrets     []byte
			installedAt time.Time
			updatedAt   time.Time
		)
		err := rows.Scan(
			&p.ID, &p.Type, &p.DisplayName, &p.Image, &p.Version, &p.GRPCAddr, &frontendURL,
			&config, &secrets, &p.Enabled, &installedAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan plugin: %w", err)
		}
		p.FrontendURL = frontendURL.String
		p.Config = json.RawMessage(config)
		p.Secrets = json.RawMessage(secrets)
		p.InstalledAt = installedAt.UTC()
		p.UpdatedAt = updatedAt.UTC()
		p.Status = plugindomain.PluginStatusUnknown
		out = append(out, &p)
	}
	return out, rows.Err()
}
