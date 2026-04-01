// Package application holds the PluginManager — the central coordinator for
// plugin lifecycle: deploy, stop, configure, health-check, and gRPC routing.
package application

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	pluginsv1 "github.com/kleffio/plugin-sdk/v1"
	grpcpool "github.com/kleffio/platform/internal/core/plugins/adapters/grpc"
	"github.com/kleffio/platform/internal/core/plugins/domain"
	"github.com/kleffio/platform/internal/core/plugins/ports"
	"github.com/kleffio/platform/internal/shared/runtime"
)

const (
	activeIDPSettingKey = "active_idp_plugin"
	grpcPort            = 50051
	healthCheckInterval = 30 * time.Second
	maxRestartAttempts  = 3
)

// Manager implements ports.PluginManager.
type Manager struct {
	store     ports.PluginStore
	registry  ports.PluginRegistry
	rt        runtime.RuntimeAdapter
	pool      *grpcpool.Pool
	secretKey []byte // 32-byte AES-256 key
	logger    *slog.Logger

	mu           sync.RWMutex
	statuses     map[string]domain.PluginStatus
	restarts     map[string]int
	capabilities map[string]map[string]bool // plugin ID → set of declared capabilities
	routes       []pluginRoute              // flat list of all declared plugin HTTP routes
}

// pluginRoute is one entry in the route registry.
type pluginRoute struct {
	pluginID string
	method   string // HTTP method or "*"
	path     string // exact or prefix ending in "*"
	public   bool
}

// New creates a PluginManager and starts the background health-check loop.
// secretKey is used for AES-256-GCM encryption of plugin secrets; it must be
// exactly 32 bytes (derive with SHA-256 from SECRET_KEY env var).
func New(
	store ports.PluginStore,
	registry ports.PluginRegistry,
	rt runtime.RuntimeAdapter,
	secretKey []byte,
	logger *slog.Logger,
) *Manager {
	m := &Manager{
		store:        store,
		registry:     registry,
		rt:           rt,
		pool:         grpcpool.NewPool(),
		secretKey:    secretKey,
		logger:       logger,
		statuses:     make(map[string]domain.PluginStatus),
		restarts:     make(map[string]int),
		capabilities: make(map[string]map[string]bool),
	}
	return m
}

// Start loads all enabled plugins from the store, ensures their containers are
// running, and starts the background health-check goroutine.
// Call this after the DB is ready.
func (m *Manager) Start(ctx context.Context) error {
	plugins, err := m.store.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("plugin manager: load plugins: %w", err)
	}

	for _, p := range plugins {
		if !p.Enabled {
			m.setStatus(p.ID, domain.PluginStatusDisabled)
			continue
		}
		// Frontend-only plugins have no backend container — mark running immediately.
		if p.GRPCAddr == "" {
			m.setStatus(p.ID, domain.PluginStatusRunning)
			continue
		}
		if err := m.ensureRunning(ctx, p); err != nil {
			m.logger.Warn("plugin manager: startup: failed to start plugin",
				"id", p.ID, "error", err)
			m.setStatus(p.ID, domain.PluginStatusError)
		}
	}

	go m.healthLoop(ctx)
	return nil
}

var _ ports.PluginManager = (*Manager)(nil)

// Install deploys a new plugin container, persists config, and dials gRPC.
func (m *Manager) Install(ctx context.Context, manifest *domain.CatalogManifest, config map[string]string) (*domain.Plugin, error) {
	// Separate plain config from secrets.
	plainCfg := map[string]string{}
	secretCfg := map[string]string{}
	for _, field := range manifest.Config {
		val, ok := config[field.Key]
		if !ok {
			continue
		}
		if field.Type == "secret" {
			secretCfg[field.Key] = val
		} else {
			plainCfg[field.Key] = val
		}
	}

	configJSON, err := json.Marshal(plainCfg)
	if err != nil {
		return nil, fmt.Errorf("install plugin: marshal config: %w", err)
	}
	secretsJSON, err := m.encryptSecrets(secretCfg)
	if err != nil {
		return nil, fmt.Errorf("install plugin: encrypt secrets: %w", err)
	}

	grpcAddr := fmt.Sprintf("kleff-%s:%d", manifest.ID, grpcPort)

	p := &domain.Plugin{
		ID:          manifest.ID,
		Type:        manifest.Type,
		DisplayName: manifest.Name,
		Image:       fmt.Sprintf("%s:%s", manifest.Image, manifest.Version),
		Version:     manifest.Version,
		GRPCAddr:    grpcAddr,
		Config:      configJSON,
		Secrets:     secretsJSON,
		Enabled:     true,
		Status:      domain.PluginStatusInstalling,
		InstalledAt: time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := m.store.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("install plugin: save: %w", err)
	}

	m.setStatus(p.ID, domain.PluginStatusInstalling)

	// Deploy companion containers declared in the manifest (e.g. Keycloak server).
	if err := m.deployCompanions(ctx, manifest, config); err != nil {
		m.setStatus(p.ID, domain.PluginStatusError)
		return nil, fmt.Errorf("install plugin: deploy companions: %w", err)
	}

	spec := m.buildContainerSpec(p, config)
	if err := m.rt.Deploy(ctx, spec); err != nil {
		m.setStatus(p.ID, domain.PluginStatusError)
		return nil, fmt.Errorf("install plugin: deploy: %w", err)
	}

	if err := m.pool.Dial(ctx, p.ID, grpcAddr); err != nil {
		m.setStatus(p.ID, domain.PluginStatusError)
		return nil, fmt.Errorf("install plugin: dial gRPC: %w", err)
	}

	m.setStatus(p.ID, domain.PluginStatusRunning)
	p.Status = domain.PluginStatusRunning
	m.logger.Info("plugin installed", "id", p.ID, "image", p.Image)
	return p, nil
}

// Remove stops the container, removes the DB record, and closes the gRPC connection.
func (m *Manager) Remove(ctx context.Context, pluginID string) error {
	// Refuse to remove the active IDP — there must always be one active IDP.
	if activeID, _ := m.store.GetSetting(ctx, activeIDPSettingKey); activeID == pluginID {
		return fmt.Errorf("remove plugin: %q is the active IDP; activate a different IDP plugin first", pluginID)
	}

	m.setStatus(pluginID, domain.PluginStatusRemoving)

	_ = m.pool.Close(pluginID)
	containerID := "kleff-" + pluginID
	if err := m.rt.Remove(ctx, containerID); err != nil {
		m.logger.Warn("plugin remove: container removal failed", "id", pluginID, "error", err)
	}

	// Best-effort removal of companion containers.
	if manifest, err := m.registry.GetManifest(ctx, pluginID); err == nil && manifest != nil {
		m.removeCompanions(ctx, manifest)
	}

	if err := m.store.Delete(ctx, pluginID); err != nil {
		return fmt.Errorf("remove plugin: delete record: %w", err)
	}

	m.mu.Lock()
	delete(m.statuses, pluginID)
	delete(m.restarts, pluginID)
	delete(m.capabilities, pluginID)
	m.mu.Unlock()

	m.logger.Info("plugin removed", "id", pluginID)
	return nil
}

// Enable starts the container and re-opens the gRPC connection.
func (m *Manager) Enable(ctx context.Context, pluginID string) error {
	p, err := m.store.FindByID(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("enable plugin: find: %w", err)
	}

	p.Enabled = true
	p.UpdatedAt = time.Now().UTC()
	if err := m.store.Save(ctx, p); err != nil {
		return fmt.Errorf("enable plugin: save: %w", err)
	}

	return m.ensureRunning(ctx, p)
}

// Disable stops the container and closes the gRPC connection.
func (m *Manager) Disable(ctx context.Context, pluginID string) error {
	// Refuse to disable the active IDP — there must always be one active IDP.
	if activeID, _ := m.store.GetSetting(ctx, activeIDPSettingKey); activeID == pluginID {
		return fmt.Errorf("disable plugin: %q is the active IDP; activate a different IDP plugin first", pluginID)
	}

	p, err := m.store.FindByID(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("disable plugin: find: %w", err)
	}

	_ = m.pool.Close(pluginID)

	containerID := "kleff-" + pluginID
	if err := m.rt.Stop(ctx, containerID); err != nil {
		m.logger.Warn("plugin disable: stop failed", "id", pluginID, "error", err)
	}

	p.Enabled = false
	p.UpdatedAt = time.Now().UTC()
	if err := m.store.Save(ctx, p); err != nil {
		return fmt.Errorf("disable plugin: save: %w", err)
	}

	m.setStatus(pluginID, domain.PluginStatusDisabled)
	m.clearCapabilities(pluginID)
	return nil
}

// Reconfigure restarts the container with updated config.
func (m *Manager) Reconfigure(ctx context.Context, pluginID string, config map[string]string) error {
	p, err := m.store.FindByID(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("reconfigure plugin: find: %w", err)
	}

	// Fetch manifest to know which fields are secrets.
	manifest, err := m.registry.GetManifest(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("reconfigure plugin: get manifest: %w", err)
	}

	plainCfg := map[string]string{}
	secretCfg := map[string]string{}
	if manifest != nil {
		for _, field := range manifest.Config {
			if val, ok := config[field.Key]; ok {
				if field.Type == "secret" {
					secretCfg[field.Key] = val
				} else {
					plainCfg[field.Key] = val
				}
			}
		}
	} else {
		plainCfg = config
	}

	configJSON, _ := json.Marshal(plainCfg)
	secretsJSON, err := m.encryptSecrets(secretCfg)
	if err != nil {
		return fmt.Errorf("reconfigure plugin: encrypt secrets: %w", err)
	}

	p.Config = configJSON
	p.Secrets = secretsJSON
	p.UpdatedAt = time.Now().UTC()

	if err := m.store.Save(ctx, p); err != nil {
		return fmt.Errorf("reconfigure plugin: save: %w", err)
	}

	// Restart container with new env vars.
	_ = m.pool.Close(pluginID)
	spec := m.buildContainerSpec(p, config)
	if err := m.rt.Deploy(ctx, spec); err != nil {
		return fmt.Errorf("reconfigure plugin: redeploy: %w", err)
	}

	return m.pool.Dial(ctx, p.ID, p.GRPCAddr)
}

// GetPlugin returns the persisted plugin with its current in-memory status.
func (m *Manager) GetPlugin(ctx context.Context, pluginID string) (*domain.Plugin, error) {
	p, err := m.store.FindByID(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	p.Status = m.getStatus(p.ID)
	return p, nil
}

// ListPlugins returns all installed plugins with in-memory statuses.
func (m *Manager) ListPlugins(ctx context.Context) ([]*domain.Plugin, error) {
	plugins, err := m.store.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plugins {
		p.Status = m.getStatus(p.ID)
	}
	return plugins, nil
}

// ── Identity (auth) ───────────────────────────────────────────────────────────

func (m *Manager) getActiveIDP(ctx context.Context) (pluginsv1.IdentityPluginClient, error) {
	activeID, err := m.store.GetSetting(ctx, activeIDPSettingKey)
	if err != nil || activeID == "" {
		return nil, err
	}
	return m.pool.IDPClient(activeID)
}

func (m *Manager) SetActiveIDP(ctx context.Context, pluginID string) error {
	return m.store.SetSetting(ctx, activeIDPSettingKey, pluginID)
}

func (m *Manager) ValidateToken(ctx context.Context, token string) (*pluginsv1.TokenClaims, error) {
	idp, err := m.getActiveIDP(ctx)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if idp == nil {
		return nil, fmt.Errorf("no active IDP plugin")
	}
	resp, err := idp.ValidateToken(ctx, &pluginsv1.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%s", resp.Error.Message)
	}
	return resp.Claims, nil
}

func (m *Manager) Login(ctx context.Context, username, password string) (*pluginsv1.TokenSet, error) {
	idp, err := m.getActiveIDP(ctx)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if idp == nil {
		return nil, fmt.Errorf("no active IDP plugin")
	}
	resp, err := idp.Login(ctx, &pluginsv1.LoginRequest{Username: username, Password: password})
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Token, nil
}

func (m *Manager) Register(ctx context.Context, req *pluginsv1.RegisterRequest) (string, error) {
	idp, err := m.getActiveIDP(ctx)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	if idp == nil {
		return "", fmt.Errorf("no active IDP plugin")
	}
	resp, err := idp.Register(ctx, req)
	if err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	if resp.Error != nil {
		return "", resp.Error
	}
	return resp.UserID, nil
}

func (m *Manager) GetOIDCConfig(ctx context.Context) (*pluginsv1.OIDCConfig, error) {
	idp, err := m.getActiveIDP(ctx)
	if err != nil || idp == nil {
		return nil, err
	}
	resp, err := idp.GetOIDCConfig(ctx, &pluginsv1.GetOIDCConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("get OIDC config: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("%s", resp.Error.Message)
	}
	return resp.Config, nil
}

func (m *Manager) RefreshToken(ctx context.Context, refreshToken string) (*pluginsv1.TokenSet, error) {
	idp, err := m.getActiveIDP(ctx)
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if idp == nil {
		return nil, fmt.Errorf("no active IDP plugin")
	}
	resp, err := idp.RefreshToken(ctx, &pluginsv1.RefreshTokenRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Token, nil
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func (m *Manager) ensureRunning(ctx context.Context, p *domain.Plugin) error {
	// Ensure companion containers (e.g. Keycloak server) are running first.
	if manifest, err := m.registry.GetManifest(ctx, p.ID); err == nil && manifest != nil {
		secrets, _ := m.decryptSecrets(p.Secrets)
		cfg := mergeConfig(p.Config, secrets)
		if err := m.deployCompanions(ctx, manifest, cfg); err != nil {
			m.logger.Warn("ensureRunning: companion deploy failed", "plugin", p.ID, "error", err)
		}
	}

	containerID := "kleff-" + p.ID
	st, err := m.rt.Status(ctx, containerID)
	if err != nil {
		return fmt.Errorf("status check: %w", err)
	}

	if st.State != runtime.StateRunning {
		// Decode secrets for env injection.
		secrets, _ := m.decryptSecrets(p.Secrets)
		allConfig := mergeConfig(p.Config, secrets)

		spec := m.buildContainerSpec(p, allConfig)
		if err := m.rt.Deploy(ctx, spec); err != nil {
			return fmt.Errorf("deploy: %w", err)
		}
	}

	if !m.pool.HasConnection(p.ID) {
		if err := m.pool.Dial(ctx, p.ID, p.GRPCAddr); err != nil {
			return fmt.Errorf("dial gRPC: %w", err)
		}
	}

	m.setStatus(p.ID, domain.PluginStatusRunning)
	m.discoverCapabilities(context.Background(), p.ID)
	return nil
}

func (m *Manager) buildContainerSpec(p *domain.Plugin, config map[string]string) runtime.ContainerSpec {
	env := make(map[string]string, len(config)+2)
	for k, v := range config {
		if v != "" {
			env[k] = v
		}
	}
	env["PLUGIN_ID"] = p.ID
	env["PLUGIN_PORT"] = fmt.Sprintf("%d", grpcPort)

	labels := map[string]string{
		"kleff.io/managed":   "true",
		"kleff.io/plugin-id": p.ID,
		"kleff.io/type":      p.Type,
	}

	return runtime.ContainerSpec{
		ID:    "kleff-" + p.ID,
		Image: p.Image,
		Env:   env,
		Ports: []runtime.PortMapping{
			{ContainerPort: grpcPort, Protocol: "tcp"},
		},
		Labels:        labels,
		RestartPolicy: runtime.RestartAlways,
	}
}

// healthLoop runs the 30-second health-check background goroutine.
func (m *Manager) healthLoop(ctx context.Context) {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runHealthChecks(ctx)
		}
	}
}

func (m *Manager) runHealthChecks(ctx context.Context) {
	plugins, err := m.store.ListAll(ctx)
	if err != nil {
		m.logger.Error("health check: list plugins", "error", err)
		return
	}

	for _, p := range plugins {
		if !p.Enabled {
			continue
		}
		m.checkPlugin(ctx, p)
	}
}

func (m *Manager) checkPlugin(ctx context.Context, p *domain.Plugin) {
	// Frontend-only plugins have no backend container — always healthy.
	if p.GRPCAddr == "" {
		m.setStatus(p.ID, domain.PluginStatusRunning)
		return
	}
	containerID := "kleff-" + p.ID
	st, err := m.rt.Status(ctx, containerID)
	if err != nil || st.State != runtime.StateRunning {
		attempts := m.getRestarts(p.ID)
		if attempts >= maxRestartAttempts {
			m.setStatus(p.ID, domain.PluginStatusError)
			m.logger.Error("plugin health: max restart attempts reached", "id", p.ID)
			return
		}

		m.logger.Warn("plugin health: container not running, attempting restart",
			"id", p.ID, "attempts", attempts+1)

		secrets, _ := m.decryptSecrets(p.Secrets)
		allConfig := mergeConfig(p.Config, secrets)
		spec := m.buildContainerSpec(p, allConfig)

		if restartErr := m.rt.Deploy(ctx, spec); restartErr != nil {
			m.incrementRestarts(p.ID)
			m.setStatus(p.ID, domain.PluginStatusError)
			return
		}
		if dialErr := m.pool.Dial(ctx, p.ID, p.GRPCAddr); dialErr != nil {
			m.incrementRestarts(p.ID)
			m.setStatus(p.ID, domain.PluginStatusError)
			return
		}
		m.resetRestarts(p.ID)
		m.setStatus(p.ID, domain.PluginStatusRunning)
		return
	}

	// Container is running — check gRPC health.
	hc, err := m.pool.HealthClient(p.ID)
	if err != nil {
		if dialErr := m.pool.Dial(ctx, p.ID, p.GRPCAddr); dialErr != nil {
			m.setStatus(p.ID, domain.PluginStatusError)
			return
		}
		hc, err = m.pool.HealthClient(p.ID)
		if err != nil {
			m.setStatus(p.ID, domain.PluginStatusError)
			return
		}
	}

	hCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := hc.Health(hCtx, &pluginsv1.HealthRequest{})
	if err != nil {
		m.setStatus(p.ID, domain.PluginStatusError)
		return
	}

	switch resp.Status {
	case pluginsv1.HealthStatusHealthy:
		m.resetRestarts(p.ID)
		m.setStatus(p.ID, domain.PluginStatusRunning)
	case pluginsv1.HealthStatusDegraded:
		m.setStatus(p.ID, domain.PluginStatusRunning) // degraded but still serving
	default:
		m.setStatus(p.ID, domain.PluginStatusError)
	}
}

// ── Capability extension points ───────────────────────────────────────────────

// discoverCapabilities calls GetCapabilities on the plugin and caches the result.
// Non-fatal: if the call fails the plugin is treated as having no capabilities.
func (m *Manager) discoverCapabilities(ctx context.Context, id string) {
	hc, err := m.pool.HealthClient(id)
	if err != nil {
		return
	}
	capCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := hc.GetCapabilities(capCtx, &pluginsv1.GetCapabilitiesRequest{})
	if err != nil {
		m.logger.Debug("plugin capabilities: GetCapabilities failed (no capabilities assumed)",
			"plugin", id, "error", err)
		return
	}

	caps := make(map[string]bool, len(resp.Capabilities))
	for _, c := range resp.Capabilities {
		caps[c] = true
	}

	m.mu.Lock()
	m.capabilities[id] = caps
	m.mu.Unlock()

	m.logger.Debug("plugin capabilities discovered", "plugin", id, "capabilities", resp.Capabilities)

	// If the plugin owns HTTP routes, fetch and register them now.
	if caps[pluginsv1.CapabilityAPIRoutes] {
		m.discoverRoutes(ctx, id)
	}
}

func (m *Manager) clearCapabilities(id string) {
	m.mu.Lock()
	delete(m.capabilities, id)
	// Remove all routes registered by this plugin.
	filtered := m.routes[:0]
	for _, r := range m.routes {
		if r.pluginID != id {
			filtered = append(filtered, r)
		}
	}
	m.routes = filtered
	m.mu.Unlock()
}

// HasIdentityProvider reports whether any active plugin declared CapabilityIdentityProvider.
func (m *Manager) HasIdentityProvider() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for id := range m.capabilities {
		if m.capabilities[id][pluginsv1.CapabilityIdentityProvider] {
			return true
		}
	}
	return false
}

// discoverRoutes calls GetRoutes on a plugin and registers them in the route table.
func (m *Manager) discoverRoutes(ctx context.Context, id string) {
	hc, err := m.pool.HTTPPluginClient(id)
	if err != nil {
		return
	}
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := hc.GetRoutes(rCtx, &pluginsv1.GetRoutesRequest{})
	if err != nil || resp.Error != nil {
		m.logger.Warn("plugin routes: GetRoutes failed", "plugin", id, "error", err)
		return
	}

	m.mu.Lock()
	// Remove any previously registered routes for this plugin.
	filtered := m.routes[:0]
	for _, r := range m.routes {
		if r.pluginID != id {
			filtered = append(filtered, r)
		}
	}
	for _, r := range resp.Routes {
		filtered = append(filtered, pluginRoute{
			pluginID: id,
			method:   r.Method,
			path:     r.Path,
			public:   r.Public,
		})
	}
	m.routes = filtered
	m.mu.Unlock()

	m.logger.Info("plugin routes registered", "plugin", id, "count", len(resp.Routes))
}

// MatchPluginRoute returns the plugin ID and public flag for the first route
// that matches the given method and path, or ok=false if none match.
func (m *Manager) MatchPluginRoute(method, path string) (pluginID string, public bool, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, r := range m.routes {
		if !routeMethodMatches(r.method, method) {
			continue
		}
		if routePathMatches(r.path, path) {
			return r.pluginID, r.public, true
		}
	}
	return "", false, false
}

// HandlePluginRoute forwards an HTTP request to the plugin's Handle gRPC method.
func (m *Manager) HandlePluginRoute(ctx context.Context, pluginID string, req *pluginsv1.HTTPRequest) (*pluginsv1.HTTPResponse, error) {
	hc, err := m.pool.HTTPPluginClient(pluginID)
	if err != nil {
		return nil, fmt.Errorf("plugin route: no client for %q: %w", pluginID, err)
	}
	resp, err := hc.Handle(ctx, &pluginsv1.HandleHTTPRequest{Request: req})
	if err != nil {
		return nil, fmt.Errorf("plugin route: Handle gRPC: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("plugin route: %s", resp.Error.Message)
	}
	return resp.Response, nil
}

func routeMethodMatches(pattern, method string) bool {
	return pattern == "*" || pattern == method
}

func routePathMatches(pattern, path string) bool {
	if strings.HasSuffix(pattern, "*") {
		return strings.HasPrefix(path, pattern[:len(pattern)-1])
	}
	return pattern == path
}

// RunMiddleware fans out OnRequest to all plugins that declared CapabilityAPIMiddleware.
// RunMiddleware fans out OnRequest to all plugins that declared CapabilityAPIMiddleware.
// The platform has already validated the token; userID and roles are the verified identity.
// Returns a non-nil error if any plugin denies the request.
func (m *Manager) RunMiddleware(ctx context.Context, userID string, roles []string, method, path string) error {
	m.mu.RLock()
	var middlewarePlugins []string
	for id, caps := range m.capabilities {
		if caps[pluginsv1.CapabilityAPIMiddleware] {
			middlewarePlugins = append(middlewarePlugins, id)
		}
	}
	m.mu.RUnlock()

	if len(middlewarePlugins) == 0 {
		return nil
	}

	req := &pluginsv1.MiddlewareRequest{
		UserID: userID,
		Roles:  roles,
		Method: method,
		Path:   path,
	}

	for _, id := range middlewarePlugins {
		mc, err := m.pool.MiddlewareClient(id)
		if err != nil {
			m.logger.Warn("plugin middleware: no client", "plugin", id, "error", err)
			continue
		}
		resp, err := mc.OnRequest(ctx, req)
		if err != nil {
			m.logger.Warn("plugin middleware: OnRequest error", "plugin", id, "error", err)
			continue // treat gRPC errors as non-blocking
		}
		if !resp.Allow {
			msg := "forbidden by plugin"
			if resp.Error != nil && resp.Error.Message != "" {
				msg = resp.Error.Message
			}
			return fmt.Errorf("%s", msg)
		}
	}
	return nil
}

// GetUIManifests collects UIManifest from every plugin that declared CapabilityUIManifest.
func (m *Manager) GetUIManifests(ctx context.Context) ([]*pluginsv1.UIManifest, error) {
	m.mu.RLock()
	var uiPlugins []string
	for id, caps := range m.capabilities {
		if caps[pluginsv1.CapabilityUIManifest] {
			uiPlugins = append(uiPlugins, id)
		}
	}
	m.mu.RUnlock()

	var manifests []*pluginsv1.UIManifest
	for _, id := range uiPlugins {
		uc, err := m.pool.UIClient(id)
		if err != nil {
			continue
		}
		resp, err := uc.GetUIManifest(ctx, &pluginsv1.GetUIManifestRequest{})
		if err != nil || resp.Error != nil {
			m.logger.Warn("plugin UI: GetUIManifest error", "plugin", id, "error", err)
			continue
		}
		if resp.Manifest != nil {
			resp.Manifest.PluginID = id
			manifests = append(manifests, resp.Manifest)
		}
	}
	return manifests, nil
}

// ── Companion container management ───────────────────────────────────────────

// deployCompanions starts companion containers declared in the plugin manifest.
// Only deploys a companion if it is not already running (idempotent).
// pluginConfig is the merged plain+secret config for the plugin; companions
// with SkipIfEnv set are skipped when the user supplied that config key.
func (m *Manager) deployCompanions(ctx context.Context, manifest *domain.CatalogManifest, pluginConfig map[string]string) error {
	for _, c := range manifest.Companions {
		if c.SkipIfEnv != "" && pluginConfig[c.SkipIfEnv] != "" {
			m.logger.Info("companion skipped: user provided external service",
				"companion", c.ID, "plugin", manifest.ID, "key", c.SkipIfEnv)
			continue
		}

		st, _ := m.rt.Status(ctx, c.ID)
		if st.State == runtime.StateRunning {
			continue
		}

		volumes := make([]runtime.VolumeMount, 0, len(c.Volumes))
		for _, v := range c.Volumes {
			volumes = append(volumes, runtime.VolumeMount{Name: v.Name, Target: v.Target})
		}

		spec := runtime.ContainerSpec{
			ID:      c.ID,
			Image:   c.Image,
			Command: c.Command,
			Env:     c.Env,
			Volumes: volumes,
			Labels: map[string]string{
				"kleff.io/managed":   "true",
				"kleff.io/plugin-id": manifest.ID,
				"kleff.io/companion": "true",
			},
			RestartPolicy: runtime.RestartAlways,
		}
		if err := m.rt.Deploy(ctx, spec); err != nil {
			return fmt.Errorf("companion %q: %w", c.ID, err)
		}
		m.logger.Info("companion deployed", "companion", c.ID, "plugin", manifest.ID)
	}
	return nil
}

// removeCompanions stops and removes companion containers declared in the manifest.
func (m *Manager) removeCompanions(ctx context.Context, manifest *domain.CatalogManifest) {
	for _, c := range manifest.Companions {
		if err := m.rt.Remove(ctx, c.ID); err != nil {
			m.logger.Warn("companion remove failed", "companion", c.ID, "plugin", manifest.ID, "error", err)
		} else {
			m.logger.Info("companion removed", "companion", c.ID, "plugin", manifest.ID)
		}
	}
}

// ── Secret encryption (AES-256-GCM) ──────────────────────────────────────────

func (m *Manager) encryptSecrets(secrets map[string]string) ([]byte, error) {
	if len(secrets) == 0 {
		return []byte("{}"), nil
	}
	plain, err := json.Marshal(secrets)
	if err != nil {
		return nil, err
	}
	if len(m.secretKey) == 0 {
		return plain, nil // no encryption in development mode
	}

	block, err := aes.NewCipher(m.secretKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, plain, nil)

	// Store as JSON so it round-trips cleanly.
	return json.Marshal(map[string][]byte{"v1": ciphertext})
}

func (m *Manager) decryptSecrets(data []byte) (map[string]string, error) {
	if len(data) == 0 || string(data) == "{}" {
		return map[string]string{}, nil
	}
	if len(m.secretKey) == 0 {
		// Development mode: data is plaintext JSON.
		var out map[string]string
		return out, json.Unmarshal(data, &out)
	}

	var wrapped map[string][]byte
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return map[string]string{}, nil
	}
	ciphertext, ok := wrapped["v1"]
	if !ok {
		return map[string]string{}, nil
	}

	block, err := aes.NewCipher(m.secretKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var out map[string]string
	return out, json.Unmarshal(plain, &out)
}

// ── Thread-safe status helpers ────────────────────────────────────────────────

func (m *Manager) setStatus(id string, s domain.PluginStatus) {
	m.mu.Lock()
	m.statuses[id] = s
	m.mu.Unlock()
}

func (m *Manager) getStatus(id string) domain.PluginStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if s, ok := m.statuses[id]; ok {
		return s
	}
	return domain.PluginStatusUnknown
}

func (m *Manager) getRestarts(id string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.restarts[id]
}

func (m *Manager) incrementRestarts(id string) {
	m.mu.Lock()
	m.restarts[id]++
	m.mu.Unlock()
}

func (m *Manager) resetRestarts(id string) {
	m.mu.Lock()
	m.restarts[id] = 0
	m.mu.Unlock()
}

// ── Config helpers ────────────────────────────────────────────────────────────

// mergeConfig merges a persisted JSON config blob and a secrets map into a
// single string map for container env injection.
func mergeConfig(configJSON []byte, secrets map[string]string) map[string]string {
	out := make(map[string]string, len(secrets))
	for k, v := range secrets {
		out[k] = v
	}
	var plain map[string]string
	if err := json.Unmarshal(configJSON, &plain); err == nil {
		for k, v := range plain {
			out[k] = v
		}
	}
	return out
}

// DeriveSecretKey derives a 32-byte AES-256 key from the SECRET_KEY env var
// using SHA-256. Call this during bootstrap.
func DeriveSecretKey(raw string) []byte {
	if raw == "" {
		return nil
	}
	h := sha256.Sum256([]byte(raw))
	return h[:]
}
