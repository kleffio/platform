// Package registry fetches and syncs the Kleff crate catalog from the remote
// crate registry (by default, github.com/kleffio/crate-registry).
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kleffio/platform/internal/core/catalog/domain"
	"github.com/kleffio/platform/internal/core/catalog/ports"
)

const defaultRegistryBaseURL = "https://raw.githubusercontent.com/kleffio/crate-registry/main"

// CrateRegistry fetches crate/blueprint/construct definitions from a remote
// registry and upserts them into the database via CatalogRepository.
type CrateRegistry struct {
	baseURL string
	client  *http.Client
}

// New creates a CrateRegistry. baseURL defaults to the official registry if empty.
// For local development, pass a file:// URL pointing to your crate-registry checkout,
// e.g. "file:///home/user/crate-registry".
func New(baseURL string) *CrateRegistry {
	if baseURL == "" {
		baseURL = defaultRegistryBaseURL
	}
	// Trim trailing slash so path joining is consistent.
	baseURL = strings.TrimRight(baseURL, "/")
	return &CrateRegistry{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
}

// Sync fetches the full registry (index → crates → blueprints) and upserts
// everything into the provided store. blueprint.json now contains all deployment
// fields (image, env, ports, outputs) previously in construct.json, so a
// Construct is derived from each blueprint during sync.
// Errors on individual files are logged but do not abort the full sync.
func (r *CrateRegistry) Sync(ctx context.Context, store ports.CatalogRepository) error {
	// 1. Fetch index.json
	indexData, err := r.fetch(ctx, "index.json")
	if err != nil {
		return fmt.Errorf("crate registry: fetch index: %w", err)
	}

	var index crateIndex
	if err := json.Unmarshal(indexData, &index); err != nil {
		return fmt.Errorf("crate registry: parse index: %w", err)
	}

	var syncErrors []string

	for _, entry := range index.allEntries() {
		cat := entry.Category
		pathPrefix := entry.PathPrefix
		ref := entry.Ref

		// 2. Fetch and upsert crate metadata
		crateData, err := r.fetch(ctx, fmt.Sprintf("%s/%s/crate.json", pathPrefix, ref.ID))
		if err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s/%s: %v", pathPrefix, ref.ID, err))
			continue
		}

		var wc wireCrate
		if err := json.Unmarshal(crateData, &wc); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s/%s parse: %v", pathPrefix, ref.ID, err))
			continue
		}

		if err := store.UpsertCrate(ctx, wc.toDomain(cat)); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s/%s upsert: %v", pathPrefix, ref.ID, err))
			continue
		}

		// 3. Fetch blueprint.json from each version folder.
		// blueprint.json now carries both user-facing config and deployment
		// fields (image, env, ports, outputs, runtime_hints overrides).
		// A Construct is derived from it using crate-level runtime_hints as defaults.
		for _, version := range ref.Versions {
			bpData, err := r.fetch(ctx, fmt.Sprintf("%s/%s/%s/blueprint.json", pathPrefix, ref.ID, version))
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s/%s: %v", pathPrefix, ref.ID, version, err))
				continue
			}

			var wb wireBlueprint
			if err := json.Unmarshal(bpData, &wb); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s/%s parse: %v", pathPrefix, ref.ID, version, err))
				continue
			}

			// Backward compatibility: older registry entries keep deployment runtime
			// fields in construct.json. Merge them when blueprint.json omits them.
			if constructData, err := r.fetch(ctx, fmt.Sprintf("%s/%s/%s/construct.json", pathPrefix, ref.ID, version)); err == nil {
				var wc wireConstruct
				if err := json.Unmarshal(constructData, &wc); err != nil {
					syncErrors = append(syncErrors, fmt.Sprintf("construct %s/%s/%s parse: %v", pathPrefix, ref.ID, version, err))
				} else {
					wb.applyConstructFallback(wc)
				}
			}

			// Optionally fetch entrypoint.sh — not all variants need a startup script.
			var startupScript string
			if scriptData, err := r.fetch(ctx, fmt.Sprintf("%s/%s/%s/entrypoint.sh", pathPrefix, ref.ID, version)); err == nil {
				startupScript = string(scriptData)
			}

			if err := store.UpsertBlueprint(ctx, wb.toDomain(wc.RuntimeHints, startupScript)); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s/%s upsert: %v", pathPrefix, ref.ID, version, err))
			}
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("crate registry sync completed with %d error(s): %s",
			len(syncErrors), strings.Join(syncErrors, "; "))
	}
	return nil
}

// fetch retrieves a single file from the registry (HTTP or file://).
func (r *CrateRegistry) fetch(ctx context.Context, path string) ([]byte, error) {
	url := r.baseURL + "/" + path

	if strings.HasPrefix(url, "file://") {
		filePath := strings.TrimPrefix(url, "file://")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", filePath, err)
		}
		return data, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: unexpected status %d", url, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
	if err != nil {
		return nil, fmt.Errorf("read body %s: %w", url, err)
	}
	return data, nil
}

// ── Wire types (registry JSON format) ────────────────────────────────────────

// crateIndex is the top-level index.json structure.
type crateIndex struct {
	Games     []crateRef `json:"games"`
	Databases []crateRef `json:"databases"`
	Cache     []crateRef `json:"cache"`
	Storage   []crateRef `json:"storage"`
	Web       []crateRef `json:"web"`
	Apps      []crateRef `json:"apps"`
	Crates    []crateRef `json:"crates"`
}

func (idx crateIndex) allEntries() []categoryEntry {
	var out []categoryEntry
	for _, ref := range idx.Games     { out = append(out, categoryEntry{"games", "games", ref}) }
	for _, ref := range idx.Databases { out = append(out, categoryEntry{"databases", "databases", ref}) }
	for _, ref := range idx.Cache     { out = append(out, categoryEntry{"cache", "cache", ref}) }
	for _, ref := range idx.Storage   { out = append(out, categoryEntry{"storage", "storage", ref}) }
	for _, ref := range idx.Web       { out = append(out, categoryEntry{"web", "web", ref}) }
	for _, ref := range idx.Apps      { out = append(out, categoryEntry{"apps", "apps", ref}) }
	for _, ref := range idx.Crates    { out = append(out, categoryEntry{"", "crates", ref}) }
	return out
}

// categoryEntry pairs a category name with its crate reference from index.json.
type categoryEntry struct {
	Category   string
	PathPrefix string
	Ref        crateRef
}

// crateRef is an entry in index.json listing a crate's version folder names.
type crateRef struct {
	ID       string   `json:"id"`
	Versions []string `json:"versions"`
}

// wireCrate maps crate.json from the registry.
// Category is derived from the directory path, not stored in crate.json.
// RuntimeHints holds the crate-level defaults that blueprint.json may partially override.
type wireCrate struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Category     string             `json:"category"`
	Description  string             `json:"description"`
	Logo         string             `json:"logo"`
	Tags         []string           `json:"tags"`
	Official     bool               `json:"official"`
	RuntimeHints domain.RuntimeHints `json:"runtime_hints"`
}

func (w wireCrate) toDomain(category string) *domain.Crate {
	resolvedCategory := strings.TrimSpace(category)
	if resolvedCategory == "" {
		resolvedCategory = strings.TrimSpace(w.Category)
	}
	if resolvedCategory == "" {
		resolvedCategory = "games"
	}

	return &domain.Crate{
		ID:          w.ID,
		Name:        w.Name,
		Category:    resolvedCategory,
		Description: w.Description,
		Logo:        w.Logo,
		Tags:        w.Tags,
		Official:    w.Official,
	}
}

// wireBlueprintExtension holds both the user-facing fields (enabled, sources) and
// the technical deployment fields (install_method, install_path, etc.) in a single
// merged object from blueprint.json. toDomain() and toConstruct() split them apart.
type wireBlueprintExtension struct {
	// User-facing
	Enabled bool     `json:"enabled"`
	Sources []string `json:"sources"`
	// Technical / deployment
	InstallMethod   string `json:"install_method"`
	InstallPath     string `json:"install_path"`
	FileExtension   string `json:"file_extension,omitempty"`
	ConfigPath      string `json:"config_path,omitempty"`
	RequiresRestart bool   `json:"requires_restart"`
}

// wireRuntimeHintsOverride holds optional per-blueprint overrides for runtime_hints.
// Pointer fields let us detect which values are actually set vs omitted in JSON.
type wireRuntimeHintsOverride struct {
	KubernetesStrategy *string `json:"kubernetes_strategy"`
	ExposeUDP          *bool   `json:"expose_udp"`
	PersistentStorage  *bool   `json:"persistent_storage"`
	StoragePath        *string `json:"storage_path"`
	StorageGB          *int    `json:"storage_gb"`
	HealthCheckPath    *string `json:"health_check_path"`
	HealthCheckPort    *int    `json:"health_check_port"`
}

// wireBlueprint maps blueprint.json from the registry. blueprint.json now contains
// all deployment fields previously in construct.json (image, env, ports, outputs,
// runtime_hints overrides). The construct field references a shared runtime ID
// (e.g. "java-21", "steamcmd") rather than a per-game construct.
type wireBlueprint struct {
	ID           string                            `json:"id"`
	Crate        string                            `json:"crate"`
	Construct    string                            `json:"construct"`
	Name         string                            `json:"name"`
	Description  string                            `json:"description"`
	Logo         string                            `json:"logo"`
	Version      string                            `json:"version"`
	Official     bool                              `json:"official"`
	Image        string                            `json:"image"`
	Constructs   map[string]string                 `json:"constructs"`
	Env          map[string]string                 `json:"env"`
	Ports        []domain.Port                     `json:"ports"`
	RuntimeHints wireRuntimeHintsOverride          `json:"runtime_hints"`
	Outputs      []domain.Output                   `json:"outputs"`
	Config       []domain.ConfigField              `json:"config"`
	Resources    domain.Resources                  `json:"resources"`
	Extensions   map[string]wireBlueprintExtension `json:"extensions"`
}

type wireConstruct struct {
	ID           string            `json:"id"`
	Crate        string            `json:"crate"`
	Blueprint    string            `json:"blueprint"`
	Image        string            `json:"image"`
	Version      string            `json:"version"`
	Env          map[string]string `json:"env"`
	Ports        []domain.Port     `json:"ports"`
	RuntimeHints domain.RuntimeHints `json:"runtime_hints"`
	Outputs      []domain.Output   `json:"outputs"`
}

func (w *wireBlueprint) applyConstructFallback(construct wireConstruct) {
	if w.Image == "" {
		w.Image = construct.Image
	}

	if len(w.Env) == 0 && len(construct.Env) > 0 {
		w.Env = construct.Env
	}

	if len(w.Ports) == 0 && len(construct.Ports) > 0 {
		w.Ports = construct.Ports
	}

	if len(w.Outputs) == 0 && len(construct.Outputs) > 0 {
		w.Outputs = construct.Outputs
	}

	if w.RuntimeHints.KubernetesStrategy == nil {
		v := construct.RuntimeHints.KubernetesStrategy
		w.RuntimeHints.KubernetesStrategy = &v
	}
	if w.RuntimeHints.ExposeUDP == nil {
		v := construct.RuntimeHints.ExposeUDP
		w.RuntimeHints.ExposeUDP = &v
	}
	if w.RuntimeHints.PersistentStorage == nil {
		v := construct.RuntimeHints.PersistentStorage
		w.RuntimeHints.PersistentStorage = &v
	}
	if w.RuntimeHints.StoragePath == nil {
		v := construct.RuntimeHints.StoragePath
		w.RuntimeHints.StoragePath = &v
	}
	if w.RuntimeHints.StorageGB == nil {
		v := construct.RuntimeHints.StorageGB
		w.RuntimeHints.StorageGB = &v
	}
	if w.RuntimeHints.HealthCheckPath == nil {
		v := construct.RuntimeHints.HealthCheckPath
		w.RuntimeHints.HealthCheckPath = &v
	}
	if w.RuntimeHints.HealthCheckPort == nil {
		v := construct.RuntimeHints.HealthCheckPort
		w.RuntimeHints.HealthCheckPort = &v
	}
}

func (w wireBlueprint) toDomain(crateHints domain.RuntimeHints, startupScript string) *domain.Blueprint {
	var bpExts map[string]domain.BlueprintExtension
	if len(w.Extensions) > 0 {
		bpExts = make(map[string]domain.BlueprintExtension, len(w.Extensions))
		for k, v := range w.Extensions {
			bpExts[k] = domain.BlueprintExtension{
				Enabled: v.Enabled,
				Sources: v.Sources,
			}
		}
	}
	return &domain.Blueprint{
		ID:            w.ID,
		CrateID:       w.Crate,
		ConstructID:   w.Construct,
		Name:          w.Name,
		Description:   w.Description,
		Logo:          w.Logo,
		Version:       w.Version,
		Official:      w.Official,
		Image:         w.Image,
		Constructs:    w.Constructs,
		Env:           w.Env,
		Ports:         w.Ports,
		Outputs:       w.Outputs,
		RuntimeHints:  mergeRuntimeHints(crateHints, w.RuntimeHints),
		StartupScript: startupScript,
		Config:        w.Config,
		Resources:     w.Resources,
		Extensions:    bpExts,
	}
}

// mergeRuntimeHints applies blueprint-level overrides onto crate-level defaults.
// Only fields explicitly present in the blueprint's runtime_hints JSON are overridden.
func mergeRuntimeHints(base domain.RuntimeHints, override wireRuntimeHintsOverride) domain.RuntimeHints {
	result := base
	if override.KubernetesStrategy != nil {
		result.KubernetesStrategy = *override.KubernetesStrategy
	}
	if override.ExposeUDP != nil {
		result.ExposeUDP = *override.ExposeUDP
	}
	if override.PersistentStorage != nil {
		result.PersistentStorage = *override.PersistentStorage
	}
	if override.StoragePath != nil {
		result.StoragePath = *override.StoragePath
	}
	if override.StorageGB != nil {
		result.StorageGB = *override.StorageGB
	}
	if override.HealthCheckPath != nil {
		result.HealthCheckPath = *override.HealthCheckPath
	}
	if override.HealthCheckPort != nil {
		result.HealthCheckPort = *override.HealthCheckPort
	}
	return result
}
