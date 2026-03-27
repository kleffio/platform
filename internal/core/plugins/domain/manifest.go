package domain

// CatalogManifest is a plugin's entry in the remote plugin registry.
// Shape mirrors the kleff-plugin.json manifest documented in PLUGIN_SPEC.md.
type CatalogManifest struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Type            string        `json:"type"`
	Description     string        `json:"description"`
	LongDescription string        `json:"longDescription,omitempty"`
	Tags            []string      `json:"tags,omitempty"`
	Author          string        `json:"author"`
	Repo            string        `json:"repo"`
	Docs            string        `json:"docs,omitempty"`
	Image           string        `json:"image"`
	Version         string        `json:"version"`
	MinKleffVersion string        `json:"minKleffVersion,omitempty"`
	License         string        `json:"license,omitempty"`
	Verified        bool          `json:"verified"`
	Logo            string        `json:"logo,omitempty"`
	Screenshots     []string      `json:"screenshots,omitempty"`
	Config          []ConfigField `json:"config,omitempty"`
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
