package domain

import "time"

// Crate is a software category that groups related blueprints.
// Examples: minecraft, redis, postgresql.
type Crate struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Description string       `json:"description"`
	Logo        string       `json:"logo"`
	Tags        []string     `json:"tags"`
	Official    bool         `json:"official"`
	Blueprints  []*Blueprint `json:"blueprints,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Blueprint is the complete definition of a runnable service.
// It contains both what the user configures (version, players, memory, etc.)
// and the deployment details (image, env, ports, runtime hints, startup script).
type Blueprint struct {
	ID           string                        `json:"id"`
	CrateID      string                        `json:"crate_id"`
	ConstructID  string                        `json:"construct_id"`
	Name         string                        `json:"name"`
	Description  string                        `json:"description"`
	Logo         string                        `json:"logo"`
	Version      string                        `json:"version"`
	Official     bool                          `json:"official"`
	Image        string                        `json:"image"`
	Images       map[string]string             `json:"images,omitempty"`
	Env          map[string]string             `json:"env"`
	Ports        []Port                        `json:"ports"`
	Outputs      []Output                      `json:"outputs"`
	RuntimeHints RuntimeHints                  `json:"runtime_hints"`
	StartupScript string                       `json:"startup_script,omitempty"`
	Config       []ConfigField                 `json:"config"`
	Resources    Resources                     `json:"resources"`
	Extensions   map[string]BlueprintExtension `json:"extensions"`
	CreatedAt    time.Time                     `json:"created_at"`
	UpdatedAt    time.Time                     `json:"updated_at"`
}

// BlueprintExtension declares that this blueprint supports an extension type
// (e.g. plugin, mod) and which sources users can install from.
type BlueprintExtension struct {
	Enabled bool     `json:"enabled"`
	Sources []string `json:"sources"`
}

// Construct is the technical recipe for running a blueprint.
// It contains the Docker image, fixed env vars, ports, runtime hints, and
// extension install details. It is never shown directly to the user.
type Construct struct {
	ID            string                         `json:"id"`
	CrateID       string                         `json:"crate_id"`
	BlueprintID   string                         `json:"blueprint_id"`
	Image         string                         `json:"image"`
	Version       string                         `json:"version"`
	Env           map[string]string              `json:"env"`
	Ports         []Port                         `json:"ports"`
	RuntimeHints  RuntimeHints                   `json:"runtime_hints"`
	Extensions    map[string]ConstructExtension  `json:"extensions"`
	Outputs       []Output                       `json:"outputs"`
	StartupScript string                         `json:"startup_script,omitempty"`
	CreatedAt     time.Time                      `json:"created_at"`
	UpdatedAt     time.Time                      `json:"updated_at"`
}

// ConstructExtension holds the technical details for installing an extension
// (e.g. which path to drop JARs, whether to restart after install).
type ConstructExtension struct {
	InstallMethod   string `json:"install_method"` // "jar-drop", "folder-drop", "file-drop", etc.
	InstallPath     string `json:"install_path"`
	FileExtension   string `json:"file_extension,omitempty"`
	ConfigPath      string `json:"config_path,omitempty"`
	RequiresRestart bool   `json:"requires_restart"`
}

// RuntimeHints control how the daemon deploys the container.
type RuntimeHints struct {
	// KubernetesStrategy is one of "", "agones", or "statefulset".
	KubernetesStrategy string `json:"kubernetes_strategy"`

	// ExposeUDP indicates whether UDP ports need host-level exposure.
	ExposeUDP bool `json:"expose_udp"`

	// PersistentStorage indicates whether a PVC should be created.
	PersistentStorage bool `json:"persistent_storage"`

	// StoragePath is the mount path inside the container for the PVC.
	StoragePath string `json:"storage_path,omitempty"`

	// StorageGB is the requested PVC size in gigabytes.
	StorageGB int `json:"storage_gb,omitempty"`

	// HealthCheckPath and HealthCheckPort are used for HTTP health probes.
	HealthCheckPath string `json:"health_check_path"`
	HealthCheckPort int    `json:"health_check_port"`
}

// Resources defines the default resource allocation for a service.
type Resources struct {
	MemoryMB      int `json:"memory_mb"`
	CPUMillicores int `json:"cpu_millicores"`
	DiskGB        int `json:"disk_gb"`
}

// Port defines a network port the container exposes.
type Port struct {
	Name      string `json:"name"`
	Container int    `json:"container"`
	Protocol  string `json:"protocol"` // "tcp" or "udp"
	Expose    bool   `json:"expose"`
	Label     string `json:"label"`
}

// ConfigField defines one field in the deploy form shown to the user.
// The key becomes an environment variable inside the container.
type ConfigField struct {
	Key                string   `json:"key"`
	Label              string   `json:"label"`
	Description        string   `json:"description,omitempty"`
	Type               string   `json:"type"` // "string", "number", "boolean", "select", "secret"
	Options            []string `json:"options,omitempty"`
	Default            any      `json:"default,omitempty"`
	Required           bool     `json:"required"`
	AutoGenerate       bool     `json:"auto_generate,omitempty"`
	AutoGenerateLength int      `json:"auto_generate_length,omitempty"`
}

// Output is a value the service exposes after starting.
// Other services in a template can reference these.
type Output struct {
	Key           string `json:"key"`
	Description   string `json:"description"`
	ValueTemplate string `json:"value_template"`
}
