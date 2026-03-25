package ports

import "context"

// Token is a standard OIDC/OAuth2 token set returned by an identity provider.
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

// IdentityProvider is the port that abstracts over different identity providers
// (Keycloak, Authentik, Ory Hydra/Kratos, Auth0, etc.).
// Adapters that implement this interface live in adapters/idp/.
type IdentityProvider interface {
	// Login authenticates a user via the headless (password) grant and returns
	// a token set. Returns a domain.AppError with status 401 on bad credentials.
	Login(ctx context.Context, username, password string) (*Token, error)

	// Register creates a new user account in the identity provider and returns
	// the provider-assigned user ID. Returns a domain.AppError with status 409
	// on duplicate username/email.
	Register(ctx context.Context, req RegisterRequest) (userID string, err error)
}
