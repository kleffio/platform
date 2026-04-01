package registry

import "github.com/kleffio/platform/internal/core/plugins/domain"

// builtinCatalog contains plugins that are always available regardless of the
// remote registry URL. These are first-party plugins shipped with the platform.
// Builtins take precedence over any same-ID entry in the remote registry.
var builtinCatalog = []*domain.CatalogManifest{
	{
		ID:          "idp-oidc",
		Name:        "Generic OIDC",
		Type:        "idp",
		Description: "Connect any OIDC-compatible identity provider — Authentik, Ory, Okta, Auth0, Azure AD, and more.",
		LongDescription: "## Generic OIDC Identity Provider\n\nConnect Kleff to any identity provider that supports OpenID Connect.\n\n" +
			"### Compatible providers\n" +
			"- **Authentik** — set issuer to your application's OIDC provider URL\n" +
			"- **Ory Hydra / Kratos** — set issuer to your Hydra issuer URL\n" +
			"- **Okta** — set issuer to `https://your-org.okta.com/oauth2/default`\n" +
			"- **Auth0** — set issuer to `https://your-tenant.auth0.com/`\n" +
			"- **Azure AD** — set issuer to `https://login.microsoftonline.com/{tenant}/v2.0`\n" +
			"- **Google** — set issuer to `https://accounts.google.com`\n" +
			"- Any other OIDC-compliant provider\n\n" +
			"### Login modes\n" +
			"- **redirect** *(recommended)* — users are sent to your IDP's login page. Works with every provider.\n" +
			"- **headless** — Kleff shows its own login form and forwards credentials via Resource Owner Password Credentials (ROPC). " +
			"Only enable if your provider supports the ROPC grant.\n\n" +
			"### User registration\n" +
			"This plugin does not create users — manage accounts through your identity provider's admin interface.",
		Tags:            []string{"oidc", "oauth2", "sso", "enterprise", "external"},
		Author:          "Kleff",
		Repo:            "https://github.com/kleffio/plugins/idp-oidc",
		Docs:            "https://docs.kleff.io/plugins/idp-oidc",
		Image:           "ghcr.io/kleffio/idp-oidc",
		Version:         "1.0.0",
		MinKleffVersion: "0.5.0",
		License:         "MIT",
		Verified:        true,
		Config: []domain.ConfigField{
			{
				Key:         "OIDC_ISSUER",
				Label:       "Issuer URL",
				Description: "Your identity provider's OIDC issuer URL. The plugin fetches configuration from {issuer}/.well-known/openid-configuration.",
				Type:        "url",
				Required:    true,
			},
			{
				Key:         "OIDC_CLIENT_ID",
				Label:       "Client ID",
				Description: "The client ID of your OIDC application.",
				Type:        "string",
				Required:    true,
			},
			{
				Key:         "OIDC_CLIENT_SECRET",
				Label:       "Client Secret",
				Description: "The client secret. Leave blank for public clients.",
				Type:        "secret",
				Required:    false,
			},
			{
				Key:         "AUTH_MODE",
				Label:       "Login Mode",
				Description: "redirect — users log in via your IDP's own login page (recommended). headless — Kleff shows its own form and forwards credentials via ROPC (only enable if your IDP supports it).",
				Type:        "select",
				Options:     []string{"redirect", "headless"},
				Required:    false,
				Default:     "redirect",
			},
		},
	},
}
