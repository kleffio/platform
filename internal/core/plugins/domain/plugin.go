// Package domain holds the core plugin domain types.
package domain

import (
	"encoding/json"
	"time"
)

// PluginStatus represents the runtime lifecycle state of a plugin.
type PluginStatus string

const (
	PluginStatusInstalling PluginStatus = "installing"
	PluginStatusRunning    PluginStatus = "running"
	PluginStatusDisabled   PluginStatus = "disabled"
	PluginStatusRemoving   PluginStatus = "removing"
	PluginStatusError      PluginStatus = "error"
	PluginStatusUnknown    PluginStatus = "unknown"
)

// Plugin is a fully installed, persisted plugin instance.
// One record per installed plugin; the primary key is the plugin manifest ID
// (e.g. "idp-keycloak"), not a surrogate.
type Plugin struct {
	// ID is the plugin manifest id, e.g. "idp-keycloak".
	ID string

	// Type is the plugin capability category, e.g. "idp".
	Type string

	// DisplayName is the human-readable label.
	DisplayName string

	// Image is the Docker image reference, e.g. "ghcr.io/kleff/idp-keycloak:1.0.0".
	Image string

	// Version is the installed version string, e.g. "1.0.0".
	Version string

	// GRPCAddr is the host:port the platform dials to reach the plugin container.
	// e.g. "kleff-idp-keycloak:50051"
	GRPCAddr string

	// Config holds non-secret configuration values as a JSON blob.
	// Values are env-var key → value, e.g. {"KEYCLOAK_URL": "http://keycloak:8080"}.
	Config json.RawMessage

	// Secrets holds secret configuration values as an AES-256-GCM encrypted JSON blob.
	// Never returned in API responses.
	Secrets json.RawMessage

	// Enabled controls whether this plugin is active.
	Enabled bool

	// Status is the in-memory runtime status (not persisted).
	Status PluginStatus

	InstalledAt time.Time
	UpdatedAt   time.Time
}
