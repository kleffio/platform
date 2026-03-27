// Package plugin is the compile-time plugin catalog and runtime factory.
//
// Each provider self-registers in its package init() by calling one of the
// Register* functions. The bootstrap layer blank-imports the adapter packages
// to trigger registration, then uses the factory functions here to instantiate
// the correct provider from the config stored in the database.
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
)

// ── Shared types ──────────────────────────────────────────────────────────────

// Token is a standard OIDC/OAuth2 token set returned by an auth provider.
type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope,omitempty"`
}

// RegisterRequest carries the data needed to create a new user account.
type RegisterRequest struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

// VerifyResult holds the verified claims extracted from a bearer token.
// Returned by AuthProvider.Verify and consumed by the auth middleware.
type VerifyResult struct {
	Subject string   // OIDC sub
	Email   string
	Roles   []string
	OrgID   string
}

// ── Plugin types ──────────────────────────────────────────────────────────────

// Type identifies the capability category of a plugin.
type Type string

const (
	// TypeAuthProvider plugins handle user authentication, registration,
	// and bearer token verification against an external identity provider.
	TypeAuthProvider Type = "auth_provider"
)

// Descriptor is the compile-time metadata for a plugin provider.
// Registered via the Register* functions; read by the marketplace API to
// show operators what providers are available and what config they require.
type Descriptor struct {
	// Provider is the unique machine-readable identifier, e.g. "keycloak".
	Provider string

	// DisplayName is the human-readable label shown in the marketplace UI.
	DisplayName string

	// Description explains what this plugin does.
	Description string

	// Type is the plugin capability category.
	Type Type

	// ConfigSchema is a JSON Schema document describing the configuration fields.
	// The panel renders a dynamic form from this schema.
	ConfigSchema json.RawMessage
}

// ── AuthProvider ──────────────────────────────────────────────────────────────

// AuthProvider is the single interface satisfied by all auth provider plugins.
// It is responsible for the full auth lifecycle:
//   - Login / Register: headless credential operations against the IDP
//   - Verify: bearer token validation for incoming API requests
//
// Each adapter in core/identity/adapters/idp/ implements this interface and
// self-registers in its package init() function.
type AuthProvider interface {
	// Login authenticates a user and returns an OIDC/OAuth2 token set.
	Login(ctx context.Context, username, password string) (*Token, error)

	// Register creates a new user account in the identity provider.
	// Returns the provider-assigned user ID.
	Register(ctx context.Context, req RegisterRequest) (userID string, err error)

	// Verify validates a raw bearer token and returns its claims.
	// This is called by the RequireAuth middleware on every authenticated request.
	// The implementation is provider-specific: JWKS validation for JWT-based
	// providers (Keycloak, Auth0, Authentik, generic OIDC) and session
	// introspection for session-based providers (Ory Kratos).
	Verify(ctx context.Context, rawToken string) (*VerifyResult, error)
}

// AuthProviderFactory constructs an AuthProvider from a JSON config blob.
type AuthProviderFactory func(configJSON json.RawMessage) (AuthProvider, error)

type authProviderEntry struct {
	desc    Descriptor
	factory AuthProviderFactory
}

var authProviders = map[string]authProviderEntry{}

// RegisterAuthProvider registers a provider factory in the compile-time catalog.
// Call this from the adapter package's init() function.
func RegisterAuthProvider(desc Descriptor, factory AuthProviderFactory) {
	authProviders[desc.Provider] = authProviderEntry{desc: desc, factory: factory}
}

// NewAuthProvider instantiates an AuthProvider by provider name and JSON config.
// Returns an error if the provider is unknown or the config is invalid.
func NewAuthProvider(provider string, configJSON json.RawMessage) (AuthProvider, error) {
	entry, ok := authProviders[provider]
	if !ok {
		return nil, fmt.Errorf("plugin: unknown auth provider %q — is the adapter package imported?", provider)
	}
	return entry.factory(configJSON)
}

// ListAuthProviderDescriptors returns all registered auth provider descriptors.
// Used by the marketplace API to enumerate available providers.
func ListAuthProviderDescriptors() []Descriptor {
	out := make([]Descriptor, 0, len(authProviders))
	for _, e := range authProviders {
		out = append(out, e.desc)
	}
	return out
}
