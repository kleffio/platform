// Package seed loads Blueprint YAML files from disk and seeds the in-memory
// catalog repository. All game-specific knowledge lives in these YAML files —
// none of it belongs in application code.
package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kleff/platform/internal/core/catalog/adapters/persistence"
	"github.com/kleff/platform/internal/core/catalog/domain"
)

// blueprintFile mirrors the YAML schema on disk.
type blueprintFile struct {
	ID          string            `yaml:"id"`
	CrateID     string            `yaml:"crate_id"`
	Name        string            `yaml:"name"`
	Image       string            `yaml:"image"`
	EnvDefaults map[string]string `yaml:"env_defaults"`
	Ports       []struct {
		ContainerPort int    `yaml:"container_port"`
		Protocol      string `yaml:"protocol"`
	} `yaml:"ports"`
	Extensibility []struct {
		Type              string   `yaml:"type"`
		Label             string   `yaml:"label"`
		Sources           []string `yaml:"sources"`
		GameVersionEnvKey string   `yaml:"game_version_env_key"`
		InstallPath       string   `yaml:"install_path"`
		FileExtension     string   `yaml:"file_extension"`
	} `yaml:"extensibility"`
}

// LoadDir reads all *.yaml files in dir, converts them to domain objects,
// and registers them in repo. Crates are derived from unique crate_id values.
func LoadDir(ctx context.Context, dir string, repo *persistence.MemoryRepository) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read blueprints dir %q: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %q: %w", path, err)
		}

		var f blueprintFile
		if err := yaml.Unmarshal(data, &f); err != nil {
			return fmt.Errorf("parse %q: %w", path, err)
		}

		bp := toDomain(f)
		repo.AddBlueprint(bp)

		// Register the crate if we haven't seen it yet.
		if _, err := repo.GetCrate(ctx, bp.CrateID); err != nil {
			repo.AddCrate(&domain.Crate{
				ID:   bp.CrateID,
				Name: titleCase(bp.CrateID),
			})
		}
	}

	return nil
}

func toDomain(f blueprintFile) *domain.Blueprint {
	var ports []domain.Port
	for _, p := range f.Ports {
		ports = append(ports, domain.Port{
			ContainerPort: p.ContainerPort,
			Protocol:      strings.ToUpper(p.Protocol),
		})
	}

	var slots []domain.ExtensibilitySlot
	for _, s := range f.Extensibility {
		var sources []domain.SourceEntry
		for _, src := range s.Sources {
			sources = append(sources, domain.SourceEntry(src))
		}
		slots = append(slots, domain.ExtensibilitySlot{
			Type:              domain.ExtensionType(s.Type),
			Label:             s.Label,
			Sources:           sources,
			GameVersionEnvKey: s.GameVersionEnvKey,
			InstallPath:       s.InstallPath,
			FileExtension:     s.FileExtension,
		})
	}

	return &domain.Blueprint{
		ID:            f.ID,
		CrateID:       f.CrateID,
		Name:          f.Name,
		Image:         f.Image,
		EnvDefaults:   f.EnvDefaults,
		Ports:         ports,
		Extensibility: slots,
	}
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
