package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/plugins/domain"
)

// PluginRegistry fetches and caches the remote plugin catalog.
type PluginRegistry interface {
	// ListCatalog returns all plugins in the cached catalog.
	// Fetches from the remote URL on first call and caches with the configured TTL.
	ListCatalog(ctx context.Context) ([]*domain.CatalogManifest, error)

	// GetManifest returns the catalog entry for the given plugin ID.
	// Returns nil, nil if not found.
	GetManifest(ctx context.Context, pluginID string) (*domain.CatalogManifest, error)

	// Refresh forces a re-fetch of the catalog from the remote registry.
	Refresh(ctx context.Context) error

	// CachedAt returns the timestamp of the last successful catalog fetch.
	// Returns the zero time if the catalog has never been fetched.
	CachedAt() string // RFC3339 or ""
}
