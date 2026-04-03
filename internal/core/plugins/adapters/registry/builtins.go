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
		Tags:            []string{"oidc", "oauth2", "sso", "enterprise", "external", "backend", "identity"},
		Capabilities:    []string{"identity.provider"},
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
	{
		ID:          "idp-keycloak",
		Name:        "Keycloak",
		Type:        "idp",
		Description: "Red Hat Keycloak identity provider. SSO, MFA, and self-hosted user management.",
		LongDescription: "## Keycloak Identity Provider\n\nConnect Kleff to your Keycloak realm for enterprise-grade identity management.\n\n" +
			"### Features\n" +
			"- Direct Access Grant (headless login)\n" +
			"- User registration via Keycloak Admin REST API\n" +
			"- RS256 JWT verification via JWKS endpoint\n" +
			"- Multi-realm support\n\n" +
			"### Setup\n" +
			"1. Create a Keycloak realm for Kleff (e.g. `kleff`).\n" +
			"2. Create a confidential client with **Direct access grants** enabled.\n" +
			"3. Create an admin service account or use the master realm admin credentials.\n" +
			"4. Fill in the config fields below and click Install.",
		Tags:            []string{"self-hosted", "sso", "enterprise", "open-source", "frontend", "backend", "identity"},
		Capabilities:    []string{"identity.provider", "ui.manifest"},
		Author:          "Kleff",
		Repo:            "https://github.com/kleffio/plugins/idp-keycloak",
		Docs:            "https://docs.kleff.io/plugins/idp-keycloak",
		Image:           "ghcr.io/kleffio/idp-keycloak",
		Version:         "1.0.0",
		MinKleffVersion: "0.5.0",
		License:         "MIT",
		Verified:        true,
		Config: []domain.ConfigField{
			{
				Key:         "KEYCLOAK_URL",
				Label:       "Keycloak URL",
				Description: "Leave blank to use the bundled Keycloak. Set this to connect to your own existing Keycloak server instead.",
				Type:        "url",
				Required:    false,
			},
			{
				Key:         "KEYCLOAK_PUBLIC_URL",
				Label:       "Public URL",
				Description: "Browser-reachable Keycloak URL. Only needed if internal and public URLs differ (e.g. behind a reverse proxy).",
				Type:        "url",
				Required:    false,
			},
			{
				Key:         "KEYCLOAK_REALM",
				Label:       "Realm",
				Description: "Keycloak realm name.",
				Type:        "string",
				Required:    false,
				Default:     "kleff",
			},
			{
				Key:         "KEYCLOAK_CLIENT_ID",
				Label:       "Client ID",
				Description: "Client ID with Direct Access Grants enabled.",
				Type:        "string",
				Required:    false,
				Default:     "kleff-panel",
			},
			{
				Key:         "KEYCLOAK_CLIENT_SECRET",
				Label:       "Client Secret",
				Description: "Client secret for confidential clients. Leave blank for public clients.",
				Type:        "secret",
				Required:    false,
			},
			{
				Key:         "KEYCLOAK_ADMIN_USER",
				Label:       "Admin Username",
				Description: "Admin account used for user registration via the Keycloak Admin API.",
				Type:        "string",
				Required:    false,
				Default:     "admin",
			},
			{
				Key:         "KEYCLOAK_ADMIN_PASSWORD",
				Label:       "Admin Password",
				Description: "Admin password for user registration.",
				Type:        "secret",
				Required:    false,
				Default:     "admin",
			},
			{
				Key:         "AUTH_MODE",
				Label:       "Login Mode",
				Description: "headless — credentials form in Kleff panel. redirect — redirect to Keycloak login page.",
				Type:        "select",
				Options:     []string{"headless", "redirect"},
				Required:    false,
				Default:     "headless",
			},
		},
		Companions: []domain.CompanionSpec{
			{
				ID:      "keycloak",
				Image:   "quay.io/keycloak/keycloak:26.1",
				Command: []string{"start-dev"},
				Env: map[string]string{
					"KC_BOOTSTRAP_ADMIN_USERNAME": "admin",
					"KC_BOOTSTRAP_ADMIN_PASSWORD": "admin",
					"KC_HTTP_PORT":                "8080",
				},
				Volumes: []domain.CompanionVolume{
					{Name: "kleff-keycloak-data", Target: "/opt/keycloak/data"},
				},
				SkipIfEnv:    "KEYCLOAK_URL",
				InternalAddr: "http://keycloak:8080",
			},
		},
	},
}
