package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"

	// Domain module HTTP handlers
	adminhttp "github.com/kleff/platform/internal/core/admin/adapters/http"
	audithttp "github.com/kleff/platform/internal/core/audit/adapters/http"
	billinghttp "github.com/kleff/platform/internal/core/billing/adapters/http"
	deploymentshttp "github.com/kleff/platform/internal/core/deployments/adapters/http"
	identityhttp "github.com/kleff/platform/internal/core/identity/adapters/http"
	idpadapters "github.com/kleff/platform/internal/core/identity/adapters/idp"
	identitycmds "github.com/kleff/platform/internal/core/identity/application/commands"
	identityports "github.com/kleff/platform/internal/core/identity/ports"
	nodeshttp "github.com/kleff/platform/internal/core/nodes/adapters/http"
	organizationshttp "github.com/kleff/platform/internal/core/organizations/adapters/http"
	profilescmds "github.com/kleff/platform/internal/core/profiles/application/commands"
	profilesqueries "github.com/kleff/platform/internal/core/profiles/application/queries"
	profileshttp "github.com/kleff/platform/internal/core/profiles/adapters/http"
	profilespersistence "github.com/kleff/platform/internal/core/profiles/adapters/persistence"
	usagehttp "github.com/kleff/platform/internal/core/usage/adapters/http"
	"github.com/kleff/platform/internal/shared/middleware"
	oidcvalidator "github.com/kleff/platform/internal/shared/oidc"
)

// Container holds all wired-up application components.
// This is the composition root — dependencies flow in one direction, from
// infrastructure outward to the HTTP layer.
type Container struct {
	Config       *Config
	Logger       *slog.Logger
	DB           *sql.DB
	TokenVerifier middleware.TokenVerifier

	// HTTP handler groups per domain module
	AuthHandler          *identityhttp.AuthHandler
	IdentityHandler      *identityhttp.Handler
	OrganizationsHandler *organizationshttp.Handler
	DeploymentsHandler   *deploymentshttp.Handler
	NodesHandler         *nodeshttp.Handler
	BillingHandler       *billinghttp.Handler
	UsageHandler         *usagehttp.Handler
	AuditHandler         *audithttp.Handler
	AdminHandler         *adminhttp.Handler
	ProfilesHandler      *profileshttp.Handler
}

// NewContainer wires up all dependencies and returns the composition root.
func NewContainer(cfg *Config, logger *slog.Logger) (*Container, error) {
	db, err := openDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// ── Identity Provider adapter ─────────────────────────────────────────────
	// Select the adapter based on IDP_PROVIDER. New adapters (Authentik, Ory,
	// Auth0, etc.) are added here without touching any other layer.
	idp, err := buildIDPAdapter(cfg)
	if err != nil {
		return nil, fmt.Errorf("build IDP adapter: %w", err)
	}

	loginCmd := identitycmds.NewLoginHandler(idp)
	registerCmd := identitycmds.NewRegisterHandler(idp)

	// ── Profiles module ──────────────────────────────────────────────────────
	profileRepo := profilespersistence.NewPostgresProfileRepository(db)
	upsertProfile := profilescmds.NewUpsertProfileHandler(profileRepo)
	updateProfile := profilescmds.NewUpdateProfileHandler(profileRepo)
	getProfile := profilesqueries.NewGetProfileHandler(profileRepo)

	return &Container{
		Config:       cfg,
		Logger:       logger,
		DB:           db,
		TokenVerifier: oidcvalidator.NewValidator(cfg.JWKSUri),

		AuthHandler:          identityhttp.NewAuthHandler(loginCmd, registerCmd, logger),
		IdentityHandler:      identityhttp.NewHandler(logger),
		OrganizationsHandler: organizationshttp.NewHandler(logger),
		DeploymentsHandler:   deploymentshttp.NewHandler(logger),
		NodesHandler:         nodeshttp.NewHandler(logger),
		BillingHandler:       billinghttp.NewHandler(logger),
		UsageHandler:         usagehttp.NewHandler(logger),
		AuditHandler:         audithttp.NewHandler(logger),
		AdminHandler:         adminhttp.NewHandler(logger),
		ProfilesHandler:      profileshttp.NewHandler(logger, upsertProfile, updateProfile, getProfile),
	}, nil
}

// buildIDPAdapter selects and constructs the correct IdentityProvider adapter
// based on the IDP_PROVIDER environment variable. Adding a new IDP requires
// only a new adapter in adapters/idp/ and a new case below.
func buildIDPAdapter(cfg *Config) (identityports.IdentityProvider, error) {
	switch cfg.IDPProvider {
	case "keycloak", "":
		return idpadapters.NewKeycloakAdapter(idpadapters.KeycloakConfig{
			BaseURL:       cfg.KeycloakURL,
			Realm:         cfg.KeycloakRealm,
			ClientID:      cfg.KeycloakClientID,
			AdminUser:     cfg.KeycloakAdminUser,
			AdminPassword: cfg.KeycloakAdminPassword,
		}), nil
	case "auth0":
		return idpadapters.NewAuth0Adapter(idpadapters.Auth0Config{
			Domain:           cfg.Auth0Domain,
			ClientID:         cfg.Auth0ClientID,
			ClientSecret:     cfg.Auth0ClientSecret,
			Audience:         cfg.Auth0Audience,
			Connection:       cfg.Auth0Connection,
			MgmtClientID:     cfg.Auth0MgmtClientID,
			MgmtClientSecret: cfg.Auth0MgmtClientSecret,
		}), nil
	case "authentik":
		return idpadapters.NewAuthentikAdapter(idpadapters.AuthentikConfig{
			BaseURL:      cfg.AuthentikBaseURL,
			ClientID:     cfg.AuthentikClientID,
			ClientSecret: cfg.AuthentikClientSecret,
			APIToken:     cfg.AuthentikAPIToken,
			FlowSlug:     cfg.AuthentikFlowSlug,
		}), nil
	case "oidc":
		return idpadapters.NewOIDCAdapter(idpadapters.OIDCConfig{
			Issuer:        cfg.OIDCGenericIssuer,
			TokenEndpoint: cfg.OIDCGenericTokenEndpoint,
			ClientID:      cfg.OIDCGenericClientID,
			ClientSecret:  cfg.OIDCGenericClientSecret,
			Scope:         cfg.OIDCGenericScope,
		}), nil
	case "ory":
		return idpadapters.NewOryAdapter(idpadapters.OryConfig{
			KratosPublicURL: cfg.OryKratosPublicURL,
			HydraAdminURL:   cfg.OryHydraAdminURL,
			HydraClientID:   cfg.OryHydraClientID,
		}), nil
	default:
		return nil, fmt.Errorf("unknown IDP_PROVIDER %q — supported: keycloak, auth0, authentik, oidc, ory", cfg.IDPProvider)
	}
}

func openDatabase(dsn string) (*sql.DB, error) {
	// TODO: import a PostgreSQL driver (e.g. _ "github.com/jackc/pgx/v5/stdlib")
	// and replace this stub with database.Connect from internal/shared/database.
	// Until then, the binary will start but all persistence calls will no-op.
	_ = dsn
	return nil, nil
}
