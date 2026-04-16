package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/catalog/domain"
	"github.com/kleffio/platform/internal/core/catalog/ports"
)

// PostgresCatalogStore implements ports.CatalogRepository against PostgreSQL.
type PostgresCatalogStore struct {
	db *sql.DB
}

func NewPostgresCatalogStore(db *sql.DB) ports.CatalogRepository {
	return &PostgresCatalogStore{db: db}
}

// ── Crates ────────────────────────────────────────────────────────────────────

func (s *PostgresCatalogStore) ListCrates(ctx context.Context, category string) ([]*domain.Crate, error) {
	query := `
		SELECT id, name, category, description, logo, tags, official, created_at, updated_at
		FROM crates`
	args := []any{}

	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY name"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list crates: %w", err)
	}
	defer rows.Close()

	var crates []*domain.Crate
	for rows.Next() {
		c, err := scanCrate(rows)
		if err != nil {
			return nil, err
		}
		crates = append(crates, c)
	}
	return crates, rows.Err()
}

func (s *PostgresCatalogStore) GetCrate(ctx context.Context, id string) (*domain.Crate, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, category, description, logo, tags, official, created_at, updated_at
		FROM crates WHERE id = $1`, id)

	c, err := scanCrate(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("crate %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get crate: %w", err)
	}

	blueprints, err := s.ListBlueprints(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Blueprints = blueprints
	return c, nil
}

func (s *PostgresCatalogStore) UpsertCrate(ctx context.Context, c *domain.Crate) error {
	tagsJSON, err := json.Marshal(c.Tags)
	if err != nil {
		return fmt.Errorf("upsert crate: marshal tags: %w", err)
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO crates (id, name, category, description, logo, tags, official, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (id) DO UPDATE SET
			name        = EXCLUDED.name,
			category    = EXCLUDED.category,
			description = EXCLUDED.description,
			logo        = EXCLUDED.logo,
			tags        = EXCLUDED.tags,
			official    = EXCLUDED.official,
			updated_at  = NOW()`,
		c.ID, c.Name, c.Category, c.Description, c.Logo, tagsJSON, c.Official,
	)
	return err
}

// ── Blueprints ────────────────────────────────────────────────────────────────

func (s *PostgresCatalogStore) ListBlueprints(ctx context.Context, crateID string) ([]*domain.Blueprint, error) {
	query := `
		SELECT id, crate_id, construct_id, name, description, logo, version,
		       official, config, resources, extensions, image, images, env, ports, outputs, runtime_hints, startup_script, created_at, updated_at
		FROM blueprints WHERE 1=1`
	args := []any{}
	i := 1

	if crateID != "" {
		query += fmt.Sprintf(" AND crate_id = $%d", i)
		args = append(args, crateID)
	}
	query += " ORDER BY name"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list blueprints: %w", err)
	}
	defer rows.Close()

	var blueprints []*domain.Blueprint
	for rows.Next() {
		b, err := scanBlueprint(rows)
		if err != nil {
			return nil, err
		}
		blueprints = append(blueprints, b)
	}
	return blueprints, rows.Err()
}

func (s *PostgresCatalogStore) GetBlueprint(ctx context.Context, id string) (*domain.Blueprint, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, crate_id, construct_id, name, description, logo, version,
		       official, config, resources, extensions, image, images, env, ports, outputs, runtime_hints, startup_script, created_at, updated_at
		FROM blueprints WHERE id = $1`, id)

	b, err := scanBlueprint(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("blueprint %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get blueprint: %w", err)
	}
	return b, nil
}

func (s *PostgresCatalogStore) UpsertBlueprint(ctx context.Context, b *domain.Blueprint) error {
	configJSON, err := json.Marshal(b.Config)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal config: %w", err)
	}
	resourcesJSON, err := json.Marshal(b.Resources)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal resources: %w", err)
	}
	extJSON, err := json.Marshal(b.Extensions)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal extensions: %w", err)
	}
	envJSON, err := json.Marshal(b.Env)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal env: %w", err)
	}
	portsJSON, err := json.Marshal(b.Ports)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal ports: %w", err)
	}
	outputsJSON, err := json.Marshal(b.Outputs)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal outputs: %w", err)
	}
	hintsJSON, err := json.Marshal(b.RuntimeHints)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal runtime_hints: %w", err)
	}
	imagesJSON, err := json.Marshal(b.Constructs)
	if err != nil {
		return fmt.Errorf("upsert blueprint: marshal constructs: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO blueprints (id, crate_id, construct_id, name, description, logo, version, official, config, resources, extensions, image, images, env, ports, outputs, runtime_hints, startup_script, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, NOW())
		ON CONFLICT (id) DO UPDATE SET
			crate_id       = EXCLUDED.crate_id,
			construct_id   = EXCLUDED.construct_id,
			name           = EXCLUDED.name,
			description    = EXCLUDED.description,
			logo           = EXCLUDED.logo,
			version        = EXCLUDED.version,
			official       = EXCLUDED.official,
			config         = EXCLUDED.config,
			resources      = EXCLUDED.resources,
			extensions     = EXCLUDED.extensions,
			image          = EXCLUDED.image,
			images         = EXCLUDED.images,
			env            = EXCLUDED.env,
			ports          = EXCLUDED.ports,
			outputs        = EXCLUDED.outputs,
			runtime_hints  = EXCLUDED.runtime_hints,
			startup_script = EXCLUDED.startup_script,
			updated_at     = NOW()`,
		b.ID, b.CrateID, b.ConstructID, b.Name, b.Description, b.Logo, b.Version, b.Official,
		configJSON, resourcesJSON, extJSON, b.Image, imagesJSON, envJSON, portsJSON, outputsJSON, hintsJSON, b.StartupScript,
	)
	return err
}

// ── Constructs ────────────────────────────────────────────────────────────────

func (s *PostgresCatalogStore) ListConstructs(ctx context.Context, crateID, blueprintID string) ([]*domain.Construct, error) {
	query := `
		SELECT id, crate_id, blueprint_id, image, version, env, ports,
		       runtime_hints, extensions, outputs, startup_script, created_at, updated_at
		FROM constructs WHERE 1=1`
	args := []any{}
	i := 1

	if crateID != "" {
		query += fmt.Sprintf(" AND crate_id = $%d", i)
		args = append(args, crateID)
		i++
	}
	if blueprintID != "" {
		query += fmt.Sprintf(" AND blueprint_id = $%d", i)
		args = append(args, blueprintID)
	}
	query += " ORDER BY id"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list constructs: %w", err)
	}
	defer rows.Close()

	var constructs []*domain.Construct
	for rows.Next() {
		c, err := scanConstruct(rows)
		if err != nil {
			return nil, err
		}
		constructs = append(constructs, c)
	}
	return constructs, rows.Err()
}

func (s *PostgresCatalogStore) GetConstruct(ctx context.Context, id string) (*domain.Construct, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, crate_id, blueprint_id, image, version, env, ports,
		       runtime_hints, extensions, outputs, startup_script, created_at, updated_at
		FROM constructs WHERE id = $1`, id)

	c, err := scanConstruct(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("construct %q not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("get construct: %w", err)
	}
	return c, nil
}

func (s *PostgresCatalogStore) UpsertConstruct(ctx context.Context, c *domain.Construct) error {
	envJSON, err := json.Marshal(c.Env)
	if err != nil {
		return fmt.Errorf("upsert construct: marshal env: %w", err)
	}
	portsJSON, err := json.Marshal(c.Ports)
	if err != nil {
		return fmt.Errorf("upsert construct: marshal ports: %w", err)
	}
	hintsJSON, err := json.Marshal(c.RuntimeHints)
	if err != nil {
		return fmt.Errorf("upsert construct: marshal runtime_hints: %w", err)
	}
	extJSON, err := json.Marshal(c.Extensions)
	if err != nil {
		return fmt.Errorf("upsert construct: marshal extensions: %w", err)
	}
	outputsJSON, err := json.Marshal(c.Outputs)
	if err != nil {
		return fmt.Errorf("upsert construct: marshal outputs: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO constructs (id, crate_id, blueprint_id, image, version, env, ports, runtime_hints, extensions, outputs, startup_script, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW())
		ON CONFLICT (id) DO UPDATE SET
			crate_id      = EXCLUDED.crate_id,
			blueprint_id  = EXCLUDED.blueprint_id,
			image         = EXCLUDED.image,
			version       = EXCLUDED.version,
			env           = EXCLUDED.env,
			ports         = EXCLUDED.ports,
			runtime_hints = EXCLUDED.runtime_hints,
			extensions    = EXCLUDED.extensions,
			outputs        = EXCLUDED.outputs,
			startup_script = EXCLUDED.startup_script,
			updated_at     = NOW()`,
		c.ID, c.CrateID, c.BlueprintID, c.Image, c.Version,
		envJSON, portsJSON, hintsJSON, extJSON, outputsJSON, c.StartupScript,
	)
	return err
}

// ── Scanners ──────────────────────────────────────────────────────────────────

type scanner interface {
	Scan(dest ...any) error
}

func scanCrate(s scanner) (*domain.Crate, error) {
	var (
		c        domain.Crate
		tagsJSON []byte
	)
	err := s.Scan(
		&c.ID, &c.Name, &c.Category, &c.Description, &c.Logo,
		&tagsJSON, &c.Official, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tagsJSON, &c.Tags); err != nil {
		return nil, fmt.Errorf("unmarshal crate tags: %w", err)
	}
	return &c, nil
}

func scanBlueprint(s scanner) (*domain.Blueprint, error) {
	var (
		b                                                              domain.Blueprint
		configJSON, resourcesJSON, extensionsJSON                     []byte
		envJSON, portsJSON, outputsJSON, hintsJSON, imagesJSON        []byte
		createdAt, updatedAt                                          time.Time
	)

	err := s.Scan(
		&b.ID, &b.CrateID, &b.ConstructID, &b.Name, &b.Description, &b.Logo,
		&b.Version, &b.Official,
		&configJSON, &resourcesJSON, &extensionsJSON,
		&b.Image, &imagesJSON, &envJSON, &portsJSON, &outputsJSON, &hintsJSON, &b.StartupScript,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	b.CreatedAt = createdAt
	b.UpdatedAt = updatedAt

	if err := json.Unmarshal(configJSON, &b.Config); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint config: %w", err)
	}
	if err := json.Unmarshal(resourcesJSON, &b.Resources); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint resources: %w", err)
	}
	if err := json.Unmarshal(extensionsJSON, &b.Extensions); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint extensions: %w", err)
	}
	if err := json.Unmarshal(envJSON, &b.Env); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint env: %w", err)
	}
	if err := json.Unmarshal(portsJSON, &b.Ports); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint ports: %w", err)
	}
	if err := json.Unmarshal(outputsJSON, &b.Outputs); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint outputs: %w", err)
	}
	if err := json.Unmarshal(hintsJSON, &b.RuntimeHints); err != nil {
		return nil, fmt.Errorf("unmarshal blueprint runtime_hints: %w", err)
	}
	if len(imagesJSON) > 0 {
		if err := json.Unmarshal(imagesJSON, &b.Constructs); err != nil {
			return nil, fmt.Errorf("unmarshal blueprint constructs: %w", err)
		}
	}

	return &b, nil
}

func scanConstruct(s scanner) (*domain.Construct, error) {
	var (
		c                                                   domain.Construct
		envJSON, portsJSON, hintsJSON, extJSON, outputsJSON []byte
		createdAt, updatedAt                                time.Time
	)

	err := s.Scan(
		&c.ID, &c.CrateID, &c.BlueprintID, &c.Image, &c.Version,
		&envJSON, &portsJSON, &hintsJSON, &extJSON, &outputsJSON,
		&createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}

	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt

	if err := json.Unmarshal(envJSON, &c.Env); err != nil {
		return nil, fmt.Errorf("unmarshal construct env: %w", err)
	}
	if err := json.Unmarshal(portsJSON, &c.Ports); err != nil {
		return nil, fmt.Errorf("unmarshal construct ports: %w", err)
	}
	if err := json.Unmarshal(hintsJSON, &c.RuntimeHints); err != nil {
		return nil, fmt.Errorf("unmarshal construct runtime_hints: %w", err)
	}
	if err := json.Unmarshal(extJSON, &c.Extensions); err != nil {
		return nil, fmt.Errorf("unmarshal construct extensions: %w", err)
	}
	if err := json.Unmarshal(outputsJSON, &c.Outputs); err != nil {
		return nil, fmt.Errorf("unmarshal construct outputs: %w", err)
	}

	return &c, nil
}
