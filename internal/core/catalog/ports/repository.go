package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/catalog/domain"
)

// CatalogRepository is the storage interface for crates, blueprints, and constructs.
type CatalogRepository interface {
	// ── Read ──────────────────────────────────────────────────────────────────

	// ListCrates returns all crates, optionally filtered by category.
	// Blueprints are not populated on list results.
	ListCrates(ctx context.Context, category string) ([]*domain.Crate, error)

	// GetCrate returns a single crate with its blueprints populated.
	GetCrate(ctx context.Context, id string) (*domain.Crate, error)

	// ListBlueprints returns all blueprints, optionally filtered by crate.
	ListBlueprints(ctx context.Context, crateID string) ([]*domain.Blueprint, error)

	// GetBlueprint returns a single blueprint by ID.
	GetBlueprint(ctx context.Context, id string) (*domain.Blueprint, error)

	// ListConstructs returns all constructs, optionally filtered by crate or blueprint.
	ListConstructs(ctx context.Context, crateID, blueprintID string) ([]*domain.Construct, error)

	// GetConstruct returns a single construct by ID.
	GetConstruct(ctx context.Context, id string) (*domain.Construct, error)

	// ── Write (used by the registry sync) ─────────────────────────────────────

	// UpsertCrate inserts or updates a crate.
	UpsertCrate(ctx context.Context, c *domain.Crate) error

	// UpsertBlueprint inserts or updates a blueprint.
	UpsertBlueprint(ctx context.Context, b *domain.Blueprint) error

	// UpsertConstruct inserts or updates a construct.
	UpsertConstruct(ctx context.Context, c *domain.Construct) error
}
