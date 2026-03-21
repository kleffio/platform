package bootstrap

import (
	"fmt"
	"strings"

	"github.com/kleff/platform/internal/shared/config"
)

// Config holds all runtime configuration for the platform API.
// Values are sourced from environment variables at startup via LoadConfig.
type Config struct {
	// HTTPPort is the port the API server listens on (default: 8080).
	HTTPPort int

	// DatabaseURL is the Postgres connection string (required).
	DatabaseURL string

	// OIDCAuthority is the OIDC provider issuer URL used for JWT verification.
	OIDCAuthority string

	// OIDCClientID is the expected audience claim in incoming JWTs.
	OIDCClientID string

	// HydraAdminURL is the internal cluster URL for Hydra's admin API.
	// Never exposed publicly. Example: http://hydra-admin.ory.svc.cluster.local:4445
	HydraAdminURL string

	// CORSAllowedOrigins is the list of allowed CORS origins.
	// An empty list permits all origins (development mode).
	CORSAllowedOrigins []string

	// LogLevel controls the minimum log severity: debug, info, warn, error.
	LogLevel string
}

// LoadConfig reads and validates configuration from environment variables.
// Returns an error if any required variable is missing or malformed.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		HTTPPort:      config.Int("HTTP_PORT", 8080),
		DatabaseURL:   config.String("DATABASE_URL", ""),
		OIDCAuthority: config.String("OIDC_AUTHORITY", ""),
		OIDCClientID:  config.String("OIDC_CLIENT_ID", "platform"),
		HydraAdminURL: config.String("HYDRA_ADMIN_URL", ""),
		LogLevel:      config.String("LOG_LEVEL", "info"),
	}

	if raw := config.String("CORS_ALLOWED_ORIGINS", ""); raw != "" {
		cfg.CORSAllowedOrigins = splitCSV(raw)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.HydraAdminURL == "" {
		return nil, fmt.Errorf("HYDRA_ADMIN_URL is required")
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
