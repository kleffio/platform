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
	nodeshttp "github.com/kleff/platform/internal/core/nodes/adapters/http"
	organizationshttp "github.com/kleff/platform/internal/core/organizations/adapters/http"
	profilescmds "github.com/kleff/platform/internal/core/profiles/application/commands"
	profilesqueries "github.com/kleff/platform/internal/core/profiles/application/queries"
	profileshttp "github.com/kleff/platform/internal/core/profiles/adapters/http"
	profilespersistence "github.com/kleff/platform/internal/core/profiles/adapters/persistence"
	usagehttp "github.com/kleff/platform/internal/core/usage/adapters/http"
	"github.com/kleff/platform/internal/shared/hydra"
)

// Container holds all wired-up application components.
// This is the composition root — dependencies flow in one direction, from
// infrastructure outward to the HTTP layer.
type Container struct {
	Config       *Config
	Logger       *slog.Logger
	DB           *sql.DB
	Introspector *hydra.Introspector

	// HTTP handler groups per domain module
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

	// ── Profiles module ──────────────────────────────────────────────────────
	// Kratos integration: the ProfileRepository uses the OIDC subject (= Kratos
	// identity.id) as its primary key. The UpsertProfileHandler creates a default
	// row on the user's first authenticated request (lazy creation strategy).
	profileRepo := profilespersistence.NewPostgresProfileRepository(db)
	upsertProfile := profilescmds.NewUpsertProfileHandler(profileRepo)
	updateProfile := profilescmds.NewUpdateProfileHandler(profileRepo)
	getProfile := profilesqueries.NewGetProfileHandler(profileRepo)

	return &Container{
		Config:       cfg,
		Logger:       logger,
		DB:           db,
		Introspector: hydra.NewIntrospector(cfg.IntrospectURL),

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

func openDatabase(dsn string) (*sql.DB, error) {
	// TODO: import a PostgreSQL driver (e.g. _ "github.com/jackc/pgx/v5/stdlib")
	// and replace this stub with database.Connect from internal/shared/database.
	// Until then, the binary will start but all persistence calls will no-op.
	_ = dsn
	return nil, nil
}
