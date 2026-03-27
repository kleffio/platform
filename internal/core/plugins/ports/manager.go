package ports

import (
	"context"

	pluginsv1 "github.com/kleffio/plugin-sdk/v1"
	"github.com/kleffio/platform/internal/core/plugins/domain"
)

// PluginManager owns the full lifecycle of installed plugins.
// It is the single authoritative source for gRPC connections and runtime state.
type PluginManager interface {
	// Install deploys the plugin container, persists its config, and opens the
	// gRPC connection.
	Install(ctx context.Context, manifest *domain.CatalogManifest, config map[string]string) (*domain.Plugin, error)

	// Remove stops the container, removes the DB record, and closes the gRPC connection.
	Remove(ctx context.Context, pluginID string) error

	// Enable starts the container and re-opens the gRPC connection.
	Enable(ctx context.Context, pluginID string) error

	// Disable stops the container and closes the gRPC connection (config preserved).
	Disable(ctx context.Context, pluginID string) error

	// Reconfigure restarts the container with updated env vars and persists the new config.
	Reconfigure(ctx context.Context, pluginID string, config map[string]string) error

	// GetPlugin returns the persisted plugin record with its current in-memory status.
	GetPlugin(ctx context.Context, pluginID string) (*domain.Plugin, error)

	// ListPlugins returns all installed plugins with their current statuses.
	ListPlugins(ctx context.Context) ([]*domain.Plugin, error)

	// ── Identity (auth) ───────────────────────────────────────────────────────

	// SetActiveIDP designates a plugin as the active identity provider.
	SetActiveIDP(ctx context.Context, pluginID string) error

	// Login authenticates a user via the active IDP plugin.
	// Returns ErrNoIDP if no active IDP is configured.
	Login(ctx context.Context, username, password string) (*pluginsv1.TokenSet, error)

	// Register creates a new user via the active IDP plugin.
	// Returns ErrNoIDP if no active IDP is configured.
	Register(ctx context.Context, req *pluginsv1.RegisterRequest) (string, error)

	// ValidateToken verifies a bearer token via the active IDP plugin.
	// Used by RequireAuth middleware on every authenticated request.
	ValidateToken(ctx context.Context, token string) (*pluginsv1.TokenClaims, error)

	// GetOIDCConfig returns the OIDC discovery config from the active IDP plugin.
	// Returns nil, nil when no IDP is active.
	GetOIDCConfig(ctx context.Context) (*pluginsv1.OIDCConfig, error)

	// RefreshToken exchanges a refresh token for a new token set via the active IDP plugin.
	// Returns ErrNoIDP if no active IDP is configured.
	RefreshToken(ctx context.Context, refreshToken string) (*pluginsv1.TokenSet, error)

	// HasIdentityProvider reports whether at least one active plugin declared
	// the CapabilityIdentityProvider capability.
	HasIdentityProvider() bool

	// ── Plugin middleware (api.middleware capability) ──────────────────────────

	// RunMiddleware calls OnRequest on every plugin that declared CapabilityAPIMiddleware.
	// Returns a non-nil error if any plugin denies the request.
	RunMiddleware(ctx context.Context, userID string, roles []string, method, path string) error

	// ── UI aggregation (ui.manifest capability) ───────────────────────────────

	// GetUIManifests aggregates UI contributions from all active plugins.
	GetUIManifests(ctx context.Context) ([]*pluginsv1.UIManifest, error)

	// ── Plugin-owned HTTP routes (api.routes capability) ──────────────────────

	// MatchPluginRoute checks whether any plugin owns the given method+path.
	MatchPluginRoute(method, path string) (pluginID string, public bool, ok bool)

	// HandlePluginRoute forwards an HTTP request to the given plugin via gRPC.
	HandlePluginRoute(ctx context.Context, pluginID string, req *pluginsv1.HTTPRequest) (*pluginsv1.HTTPResponse, error)
}
