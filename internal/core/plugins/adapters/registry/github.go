// Package registry fetches and caches the Kleff plugin catalog from the remote
// plugin registry (by default, github.com/kleff/plugin-registry).
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/kleffio/platform/internal/core/plugins/domain"
	"github.com/kleffio/platform/internal/core/plugins/ports"
)

const defaultCatalogURL = "https://raw.githubusercontent.com/kleffio/plugin-registry/main/plugins.json"

// Registry implements ports.PluginRegistry by fetching a JSON catalog from a
// remote URL and caching the result for a configurable TTL.
type Registry struct {
	url      string
	ttl      time.Duration
	client   *http.Client
	mu       sync.RWMutex
	catalog  []*domain.CatalogManifest
	cachedAt time.Time
}

// New creates a Registry. catalogURL defaults to the official registry if empty.
// ttl controls the cache lifetime (default: 1 hour).
func New(catalogURL string, ttl time.Duration) *Registry {
	if catalogURL == "" {
		catalogURL = defaultCatalogURL
	}
	if ttl <= 0 {
		ttl = time.Hour
	}
	return &Registry{
		url:    catalogURL,
		ttl:    ttl,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

var _ ports.PluginRegistry = (*Registry)(nil)

// ListCatalog returns the cached catalog merged with builtin plugins, refreshing if stale.
// Builtin plugins always appear first and are never overridden by the remote catalog.
func (r *Registry) ListCatalog(ctx context.Context) ([]*domain.CatalogManifest, error) {
	r.mu.RLock()
	if r.catalog != nil && time.Since(r.cachedAt) < r.ttl {
		c := r.mergeWithBuiltins(r.catalog)
		r.mu.RUnlock()
		return c, nil
	}
	r.mu.RUnlock()

	if err := r.Refresh(ctx); err != nil {
		// Return builtins + stale cache if available.
		r.mu.RLock()
		c := r.catalog
		r.mu.RUnlock()
		return r.mergeWithBuiltins(c), nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mergeWithBuiltins(r.catalog), nil
}

// mergeWithBuiltins prepends builtinCatalog entries, skipping any remote entries
// that share an ID with a builtin (builtins win).
func (r *Registry) mergeWithBuiltins(remote []*domain.CatalogManifest) []*domain.CatalogManifest {
	builtinIDs := make(map[string]struct{}, len(builtinCatalog))
	for _, b := range builtinCatalog {
		builtinIDs[b.ID] = struct{}{}
	}
	merged := make([]*domain.CatalogManifest, 0, len(builtinCatalog)+len(remote))
	merged = append(merged, builtinCatalog...)
	for _, m := range remote {
		if _, isBuiltin := builtinIDs[m.ID]; !isBuiltin {
			merged = append(merged, m)
		}
	}
	return merged
}

// GetManifest returns the catalog entry for the given plugin ID, or nil, nil.
// Builtin plugins are checked first.
func (r *Registry) GetManifest(ctx context.Context, pluginID string) (*domain.CatalogManifest, error) {
	for _, b := range builtinCatalog {
		if b.ID == pluginID {
			return b, nil
		}
	}
	catalog, err := r.ListCatalog(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range catalog {
		if m.ID == pluginID {
			return m, nil
		}
	}
	return nil, nil
}

// Refresh forces a re-fetch from the remote registry.
func (r *Registry) Refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return fmt.Errorf("plugin registry: build request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("plugin registry: fetch catalog: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("plugin registry: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20)) // 10 MB max
	if err != nil {
		return fmt.Errorf("plugin registry: read body: %w", err)
	}

	var manifests []*domain.CatalogManifest
	if err := json.Unmarshal(body, &manifests); err != nil {
		return fmt.Errorf("plugin registry: parse catalog: %w", err)
	}

	r.mu.Lock()
	r.catalog = manifests
	r.cachedAt = time.Now()
	r.mu.Unlock()

	return nil
}

// CachedAt returns the RFC3339 timestamp of the last successful fetch.
func (r *Registry) CachedAt() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.cachedAt.IsZero() {
		return ""
	}
	return r.cachedAt.UTC().Format(time.RFC3339)
}
