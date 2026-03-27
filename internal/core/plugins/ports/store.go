package ports

import (
	"context"

	"github.com/kleffio/platform/internal/core/plugins/domain"
)

// PluginStore persists installed plugin records and the active-plugin settings.
// Schema: see migrations/003_create_plugins.sql and 004_create_settings.sql.
type PluginStore interface {
	// FindByID returns an installed plugin by its manifest ID.
	FindByID(ctx context.Context, id string) (*domain.Plugin, error)

	// ListAll returns every installed plugin.
	ListAll(ctx context.Context) ([]*domain.Plugin, error)

	// ListByType returns all installed plugins of the given type.
	ListByType(ctx context.Context, pluginType string) ([]*domain.Plugin, error)

	// Save upserts a plugin record (insert or update by ID).
	Save(ctx context.Context, p *domain.Plugin) error

	// Delete removes a plugin record by ID.
	Delete(ctx context.Context, id string) error

	// GetSetting returns the value of a named settings key.
	// Returns "", nil if the key does not exist.
	GetSetting(ctx context.Context, key string) (string, error)

	// SetSetting upserts a named settings key.
	SetSetting(ctx context.Context, key, value string) error
}
