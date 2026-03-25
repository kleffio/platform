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

	// IntrospectURL is the OAuth2 token introspection endpoint (RFC 7662).
	// Set OAUTH2_INTROSPECT_URL to the full URL for your identity provider:
	//   Ory Hydra:    https://hydra-admin.example.com/admin/oauth2/introspect
	//   Keycloak:     https://keycloak.example.com/realms/<realm>/protocol/openid-connect/token/introspect
	// Optional — if unset and OIDC_AUTHORITY is provided, falls back to {OIDC_AUTHORITY}/oauth2/introspect.
	IntrospectURL string

	// IntrospectClientID is the OAuth2 client ID used to authenticate introspection requests.
	// For Keycloak, this must be a confidential client. Set via INTROSPECT_CLIENT_ID.
	IntrospectClientID string

	// IntrospectClientSecret is the OAuth2 client secret for the introspect client.
	// Set via INTROSPECT_CLIENT_SECRET.
	IntrospectClientSecret string

	// CORSAllowedOrigins is the list of allowed CORS origins.
	// An empty list permits all origins (development mode).
	CORSAllowedOrigins []string

	// LogLevel controls the minimum log severity: debug, info, warn, error.
	LogLevel string

	// IDPProvider selects the identity provider adapter to load.
	// Supported values: "keycloak" (default).
	// Set via IDP_PROVIDER.
	IDPProvider string

	// KeycloakURL is the Keycloak server root URL, e.g. "http://keycloak:8080".
	// Set via KEYCLOAK_URL. Required when IDPProvider == "keycloak".
	KeycloakURL string

	// KeycloakRealm is the Keycloak realm name.
	// Set via KEYCLOAK_REALM (default: "master").
	KeycloakRealm string

	// KeycloakClientID is the OIDC client used for the Direct Access Grant.
	// The client must have "Direct access grants" enabled in Keycloak.
	// Set via KEYCLOAK_CLIENT_ID (default: "kleff-panel").
	KeycloakClientID string

	// KeycloakAdminUser is the Keycloak admin username (master realm).
	// Set via KEYCLOAK_ADMIN.
	KeycloakAdminUser string

	// KeycloakAdminPassword is the Keycloak admin password.
	// Set via KEYCLOAK_ADMIN_PASSWORD.
	KeycloakAdminPassword string

	// ── Auth0 ─────────────────────────────────────────────────────────────────

	// Auth0Domain is the Auth0 tenant domain, e.g. "dev-xxxxx.us.auth0.com".
	// Set via AUTH0_DOMAIN. Required when IDPProvider == "auth0".
	Auth0Domain string

	// Auth0ClientID is the application client ID for the password grant (ROPC).
	// Set via AUTH0_CLIENT_ID.
	Auth0ClientID string

	// Auth0ClientSecret is the application client secret.
	// Set via AUTH0_CLIENT_SECRET.
	Auth0ClientSecret string

	// Auth0Audience is the API audience/identifier.
	// Set via AUTH0_AUDIENCE.
	Auth0Audience string

	// Auth0Connection is the Auth0 database connection name.
	// Set via AUTH0_CONNECTION (default: "Username-Password-Authentication").
	Auth0Connection string

	// Auth0MgmtClientID is the M2M application client ID for the Management API.
	// Set via AUTH0_MGMT_CLIENT_ID.
	Auth0MgmtClientID string

	// Auth0MgmtClientSecret is the Management API application client secret.
	// Set via AUTH0_MGMT_CLIENT_SECRET.
	Auth0MgmtClientSecret string

	// ── Authentik ─────────────────────────────────────────────────────────────

	// AuthentikBaseURL is the Authentik server root URL, e.g. "https://authentik.example.com".
	// Set via AUTHENTIK_URL. Required when IDPProvider == "authentik".
	AuthentikBaseURL string

	// AuthentikClientID is the OAuth2 application client ID.
	// Set via AUTHENTIK_CLIENT_ID.
	AuthentikClientID string

	// AuthentikClientSecret is the OAuth2 application client secret.
	// Set via AUTHENTIK_CLIENT_SECRET.
	AuthentikClientSecret string

	// AuthentikAPIToken is a service account API token for user management.
	// Set via AUTHENTIK_API_TOKEN.
	AuthentikAPIToken string

	// AuthentikFlowSlug is the authentication flow slug.
	// Set via AUTHENTIK_FLOW_SLUG (default: "default-authentication-flow").
	AuthentikFlowSlug string

	// ── Generic OIDC ──────────────────────────────────────────────────────────

	// OIDCGenericIssuer is the issuer URL for the generic OIDC adapter.
	// Set via OIDC_GENERIC_ISSUER. Required when IDPProvider == "oidc".
	OIDCGenericIssuer string

	// OIDCGenericTokenEndpoint overrides the token endpoint URL.
	// Set via OIDC_GENERIC_TOKEN_ENDPOINT. Defaults to {issuer}/oauth2/token.
	OIDCGenericTokenEndpoint string

	// OIDCGenericClientID is the OAuth2 client ID.
	// Set via OIDC_GENERIC_CLIENT_ID.
	OIDCGenericClientID string

	// OIDCGenericClientSecret is the OAuth2 client secret (optional for public clients).
	// Set via OIDC_GENERIC_CLIENT_SECRET.
	OIDCGenericClientSecret string

	// OIDCGenericScope overrides the requested scope.
	// Set via OIDC_GENERIC_SCOPE (default: "openid profile email offline_access").
	OIDCGenericScope string

	// ── Ory (Kratos + Hydra) ──────────────────────────────────────────────────

	// OryKratosPublicURL is the root URL of the Kratos public API.
	// Set via ORY_KRATOS_PUBLIC_URL. Required when IDPProvider == "ory".
	OryKratosPublicURL string

	// OryHydraAdminURL is the Hydra admin API root URL.
	// Set via ORY_HYDRA_ADMIN_URL. Optional — when set, login returns a Hydra
	// JWT access token; otherwise the Kratos session token is returned.
	OryHydraAdminURL string

	// OryHydraClientID is the OAuth2 client ID used for Hydra token creation.
	// Set via ORY_HYDRA_CLIENT_ID. Required when OryHydraAdminURL is set.
	OryHydraClientID string

	// JWKSUri is the OIDC JWKS endpoint for JWT signature verification.
	// Set via JWKS_URI. If unset, derived from OIDC_AUTHORITY as {authority}/protocol/openid-connect/certs.
	JWKSUri string
}

// LoadConfig reads and validates configuration from environment variables.
// Returns an error if any required variable is missing or malformed.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		HTTPPort:      config.Int("HTTP_PORT", 8080),
		DatabaseURL:   config.String("DATABASE_URL", ""),
		OIDCAuthority:          config.String("OIDC_AUTHORITY", ""),
		OIDCClientID:           config.String("OIDC_CLIENT_ID", "platform"),
		IntrospectURL:          config.String("OAUTH2_INTROSPECT_URL", ""),
		IntrospectClientID:     config.String("INTROSPECT_CLIENT_ID", ""),
		IntrospectClientSecret: config.String("INTROSPECT_CLIENT_SECRET", ""),
		LogLevel:               config.String("LOG_LEVEL", "info"),

		IDPProvider:           config.String("IDP_PROVIDER", "keycloak"),
		KeycloakURL:           config.String("KEYCLOAK_URL", ""),
		KeycloakRealm:         config.String("KEYCLOAK_REALM", "master"),
		KeycloakClientID:      config.String("KEYCLOAK_CLIENT_ID", "kleff-panel"),
		KeycloakAdminUser:     config.String("KEYCLOAK_ADMIN", ""),
		KeycloakAdminPassword: config.String("KEYCLOAK_ADMIN_PASSWORD", ""),

		Auth0Domain:           config.String("AUTH0_DOMAIN", ""),
		Auth0ClientID:         config.String("AUTH0_CLIENT_ID", ""),
		Auth0ClientSecret:     config.String("AUTH0_CLIENT_SECRET", ""),
		Auth0Audience:         config.String("AUTH0_AUDIENCE", ""),
		Auth0Connection:       config.String("AUTH0_CONNECTION", ""),
		Auth0MgmtClientID:     config.String("AUTH0_MGMT_CLIENT_ID", ""),
		Auth0MgmtClientSecret: config.String("AUTH0_MGMT_CLIENT_SECRET", ""),

		AuthentikBaseURL:      config.String("AUTHENTIK_URL", ""),
		AuthentikClientID:     config.String("AUTHENTIK_CLIENT_ID", ""),
		AuthentikClientSecret: config.String("AUTHENTIK_CLIENT_SECRET", ""),
		AuthentikAPIToken:     config.String("AUTHENTIK_API_TOKEN", ""),
		AuthentikFlowSlug:     config.String("AUTHENTIK_FLOW_SLUG", ""),

		OIDCGenericIssuer:        config.String("OIDC_GENERIC_ISSUER", ""),
		OIDCGenericTokenEndpoint: config.String("OIDC_GENERIC_TOKEN_ENDPOINT", ""),
		OIDCGenericClientID:      config.String("OIDC_GENERIC_CLIENT_ID", ""),
		OIDCGenericClientSecret:  config.String("OIDC_GENERIC_CLIENT_SECRET", ""),
		OIDCGenericScope:         config.String("OIDC_GENERIC_SCOPE", ""),

		OryKratosPublicURL: config.String("ORY_KRATOS_PUBLIC_URL", ""),
		OryHydraAdminURL:   config.String("ORY_HYDRA_ADMIN_URL", ""),
		OryHydraClientID:   config.String("ORY_HYDRA_CLIENT_ID", ""),

		JWKSUri: config.String("JWKS_URI", ""),
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
