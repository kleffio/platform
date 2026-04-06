package domain

// CatalogManifest is a plugin's entry in the remote plugin registry.
// Shape mirrors the kleff-plugin.json manifest documented in PLUGIN_SPEC.md.
type CatalogManifest struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Type            string          `json:"type"`
	Description     string          `json:"description"`
	LongDescription string          `json:"longDescription,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Capabilities    []string        `json:"capabilities,omitempty"`
	Author          string          `json:"author"`
	Repo            string          `json:"repo"`
	Docs            string          `json:"docs,omitempty"`
	Image           string          `json:"image"`
	Version         string          `json:"version"`
	MinKleffVersion string          `json:"minKleffVersion,omitempty"`
	License         string          `json:"license,omitempty"`
	Verified        bool            `json:"verified"`
	Logo            string          `json:"logo,omitempty"`
	Screenshots     []string        `json:"screenshots,omitempty"`
	Config          []ConfigField   `json:"config,omitempty"`
	Companions      []CompanionSpec `json:"companions,omitempty"`
}

// CompanionSpec declares a dependency container that the platform spins up
// alongside the plugin container. The companion shares the plugin's network
// and is managed (deploy/remove) together with the plugin.
type CompanionSpec struct {
	// ID is the container name on the kleff network, e.g. "keycloak".
	// Must be unique across all installed plugins.
	ID string `json:"id"`

	// Image is the Docker image reference for the companion container.
	Image string `json:"image"`

	// Command overrides the container's default CMD, e.g. ["start-dev"].
	Command []string `json:"command,omitempty"`

	// Env is a set of static environment variables injected into the companion.
	Env map[string]string `json:"env,omitempty"`

	// Ports exposes companion container ports on the host.
	Ports []CompanionPort `json:"ports,omitempty"`

	// Volumes declares named volumes mounted into the companion for persistence.
	Volumes []CompanionVolume `json:"volumes,omitempty"`

	// SkipIfEnv names a plugin config key: if the user supplied a non-empty
	// value for that key, the companion is not deployed (the user is providing
	// their own external service instead).
	SkipIfEnv string `json:"skipIfEnv,omitempty"`

	// InternalAddr is the address the plugin should use to reach this companion
	// when it is deployed (i.e. when SkipIfEnv is unset). The platform injects
	// this value as the SkipIfEnv env var so the plugin always has a valid URL.
	// Example: "http://keycloak:8080"
	InternalAddr string `json:"internalAddr,omitempty"`
}

// CompanionPort maps a container port to an optional fixed host port.
type CompanionPort struct {
	ContainerPort int    `json:"container"`
	HostPort      int    `json:"host,omitempty"` // 0 = auto-assign
	Protocol      string `json:"protocol,omitempty"` // default: "tcp"
}

// CompanionVolume maps a named Docker volume to a path inside the companion container.
type CompanionVolume struct {
	Name   string `json:"name"`   // Docker volume name, e.g. "kleff-keycloak-data"
	Target string `json:"target"` // Mount path inside container, e.g. "/opt/keycloak/data"
}

// ConfigField describes one configuration value the plugin expects.
// These are rendered as form fields in the Install/Configure modal and
// injected as environment variables into the plugin container.
type ConfigField struct {
	// Key is the environment variable name injected into the container.
	Key string `json:"key"`

	// Label is the human-readable form field label.
	Label string `json:"label"`

	// Description is shown below the input field.
	Description string `json:"description,omitempty"`

	// Type is one of: string, secret, number, boolean, select, url.
	Type string `json:"type"`

	// Required indicates the admin must fill this in before installing.
	Required bool `json:"required"`

	// Default is the pre-filled default value (optional).
	Default string `json:"default,omitempty"`

	// Options is the list of choices for type "select".
	Options []string `json:"options,omitempty"`
}
