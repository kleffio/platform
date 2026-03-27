package bootstrap

import (
	"fmt"
	"strings"

	"github.com/kleffio/platform/internal/shared/config"
)

// Config holds all runtime configuration for the platform API.
// Values are sourced from environment variables at startup via LoadConfig.
type Config struct {
	// ── Core ──────────────────────────────────────────────────────────────────

	// HTTPPort is the port the API server listens on (default: 8080).
	HTTPPort int

	// DatabaseURL is the Postgres connection string (required).
	DatabaseURL string

	// LogLevel controls the minimum log severity: debug, info, warn, error.
	LogLevel string

	// CORSAllowedOrigins is the list of allowed CORS origins.
	// An empty list permits all origins (development mode).
	CORSAllowedOrigins []string

	// ── Auth / OIDC ───────────────────────────────────────────────────────────

	// OIDCAuthority is the OIDC provider issuer URL used for JWT verification.
	OIDCAuthority string

	// OIDCClientID is the expected audience claim in incoming JWTs.
	OIDCClientID string

	// IntrospectURL is the OAuth2 token introspection endpoint (RFC 7662).
	IntrospectURL string

	// IntrospectClientID is the OAuth2 client ID for introspection requests.
	IntrospectClientID string

	// IntrospectClientSecret is the OAuth2 client secret for introspection.
	IntrospectClientSecret string

	// JWKSUri is the OIDC JWKS endpoint for JWT signature verification.
	JWKSUri string

	// ── Plugin system ─────────────────────────────────────────────────────────

	// RuntimeProvider selects the container runtime for plugin management.
	// Values: "docker" (default), "kubernetes", "manual".
	RuntimeProvider string

	// PluginRegistryURL overrides the plugin catalog URL.
	// Default: https://raw.githubusercontent.com/kleff/plugin-registry/main/plugins.json
	PluginRegistryURL string

	// PluginNetwork is the Docker network plugins are attached to (default: "kleff").
	PluginNetwork string

	// PluginGRPCPort is the port plugins listen on inside their container (default: 50051).
	PluginGRPCPort int

	// PluginRegistryTTL is the seconds between catalog re-fetches (default: 3600).
	PluginRegistryTTL int

	// PluginNamespace is the k8s namespace for plugin Deployments (default: "kleff").
	PluginNamespace string

	// SecretKey is the AES-256 key (any length; hashed with SHA-256) for
	// encrypting plugin secrets at rest. Required in production.
	SecretKey string
}

// LoadConfig reads and validates configuration from environment variables.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		HTTPPort:    config.Int("HTTP_PORT", 8080),
		DatabaseURL: config.String("DATABASE_URL", ""),
		LogLevel:    config.String("LOG_LEVEL", "info"),

		OIDCAuthority:          config.String("OIDC_AUTHORITY", ""),
		OIDCClientID:           config.String("OIDC_CLIENT_ID", "platform"),
		IntrospectURL:          config.String("OAUTH2_INTROSPECT_URL", ""),
		IntrospectClientID:     config.String("INTROSPECT_CLIENT_ID", ""),
		IntrospectClientSecret: config.String("INTROSPECT_CLIENT_SECRET", ""),
		JWKSUri:                config.String("JWKS_URI", ""),

		RuntimeProvider:   config.String("RUNTIME_PROVIDER", "docker"),
		PluginRegistryURL: config.String("PLUGIN_REGISTRY_URL", ""),
		PluginNetwork:     config.String("PLUGIN_NETWORK", "kleff"),
		PluginGRPCPort:    config.Int("PLUGIN_GRPC_PORT", 50051),
		PluginRegistryTTL: config.Int("PLUGIN_REGISTRY_TTL", 3600),
		PluginNamespace:   config.String("PLUGIN_NAMESPACE", "kleff"),
		SecretKey:         config.String("SECRET_KEY", ""),
	}

	if raw := config.String("CORS_ALLOWED_ORIGINS", ""); raw != "" {
		cfg.CORSAllowedOrigins = splitCSV(raw)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Derive introspection URL from OIDC authority if not explicitly set.
	if cfg.IntrospectURL == "" && cfg.OIDCAuthority != "" {
		cfg.IntrospectURL = strings.TrimRight(cfg.OIDCAuthority, "/") + "/oauth2/introspect"
	}

	if cfg.JWKSUri == "" && cfg.OIDCAuthority != "" {
		cfg.JWKSUri = strings.TrimRight(cfg.OIDCAuthority, "/") + "/protocol/openid-connect/certs"
	}

	return cfg, nil
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}
