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

// Sync fetches the full registry (index → crates → blueprints → constructs) and
// upserts everything into the provided store. Errors on individual files are
// logged but do not abort the full sync — partial data is better than nothing.
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

	for _, ref := range index.Crates {
		// 2. Fetch and upsert crate metadata
		crateData, err := r.fetch(ctx, fmt.Sprintf("crates/%s/crate.json", ref.ID))
		if err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s: %v", ref.ID, err))
			continue
		}

		var wc wireCrate
		if err := json.Unmarshal(crateData, &wc); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s parse: %v", ref.ID, err))
			continue
		}

		if err := store.UpsertCrate(ctx, wc.toDomain()); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("crate %s upsert: %v", ref.ID, err))
			continue
		}

		// 3. Fetch and upsert blueprints
		for _, bpID := range ref.Blueprints {
			bpData, err := r.fetch(ctx, fmt.Sprintf("crates/%s/blueprints/%s.json", ref.ID, bpID))
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s: %v", ref.ID, bpID, err))
				continue
			}

			var wb wireBlueprint
			if err := json.Unmarshal(bpData, &wb); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s parse: %v", ref.ID, bpID, err))
				continue
			}

			if err := store.UpsertBlueprint(ctx, wb.toDomain()); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("blueprint %s/%s upsert: %v", ref.ID, bpID, err))
			}
		}

		// 4. Fetch and upsert constructs
		for _, cID := range ref.Constructs {
			cData, err := r.fetch(ctx, fmt.Sprintf("crates/%s/constructs/%s.json", ref.ID, cID))
			if err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("construct %s/%s: %v", ref.ID, cID, err))
				continue
			}

			var wc wireConstruct
			if err := json.Unmarshal(cData, &wc); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("construct %s/%s parse: %v", ref.ID, cID, err))
				continue
			}

			if err := store.UpsertConstruct(ctx, wc.toDomain()); err != nil {
				syncErrors = append(syncErrors, fmt.Sprintf("construct %s/%s upsert: %v", ref.ID, cID, err))
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
	Crates []crateRef `json:"crates"`
}

// crateRef is an entry in index.json listing a crate's blueprint and construct IDs.
type crateRef struct {
	ID         string   `json:"id"`
	Blueprints []string `json:"blueprints"`
	Constructs []string `json:"constructs"`
}

// wireCrate maps crate.json from the registry.
type wireCrate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Logo        string   `json:"logo"`
	Tags        []string `json:"tags"`
	Official    bool     `json:"official"`
}

func (w wireCrate) toDomain() *domain.Crate {
	return &domain.Crate{
		ID:          w.ID,
		Name:        w.Name,
		Category:    w.Category,
		Description: w.Description,
		Logo:        w.Logo,
		Tags:        w.Tags,
		Official:    w.Official,
	}
}

// wireBlueprint maps blueprints/*.json from the registry.
// Note: "crate" and "construct" are the registry field names, not crate_id/construct_id.
type wireBlueprint struct {
	ID          string                              `json:"id"`
	Crate       string                              `json:"crate"`
	Construct   string                              `json:"construct"`
	Name        string                              `json:"name"`
	Description string                              `json:"description"`
	Logo        string                              `json:"logo"`
	Version     string                              `json:"version"`
	Official    bool                                `json:"official"`
	Config      []domain.ConfigField                `json:"config"`
	Resources   domain.Resources                    `json:"resources"`
	Extensions  map[string]domain.BlueprintExtension `json:"extensions"`
}

func (w wireBlueprint) toDomain() *domain.Blueprint {
	return &domain.Blueprint{
		ID:          w.ID,
		CrateID:     w.Crate,
		ConstructID: w.Construct,
		Name:        w.Name,
		Description: w.Description,
		Logo:        w.Logo,
		Version:     w.Version,
		Official:    w.Official,
		Config:      w.Config,
		Resources:   w.Resources,
		Extensions:  w.Extensions,
	}
}

// wireConstruct maps constructs/*.json from the registry.
type wireConstruct struct {
	ID           string                               `json:"id"`
	Crate        string                               `json:"crate"`
	Blueprint    string                               `json:"blueprint"`
	Image        string                               `json:"image"`
	Version      string                               `json:"version"`
	Env          map[string]string                    `json:"env"`
	Ports        []domain.Port                        `json:"ports"`
	RuntimeHints domain.RuntimeHints                  `json:"runtime_hints"`
	Extensions   map[string]domain.ConstructExtension `json:"extensions"`
	Outputs      []domain.Output                      `json:"outputs"`
}

func (w wireConstruct) toDomain() *domain.Construct {
	return &domain.Construct{
		ID:           w.ID,
		CrateID:      w.Crate,
		BlueprintID:  w.Blueprint,
		Image:        w.Image,
		Version:      w.Version,
		Env:          w.Env,
		Ports:        w.Ports,
		RuntimeHints: w.RuntimeHints,
		Extensions:   w.Extensions,
		Outputs:      w.Outputs,
	}
}
