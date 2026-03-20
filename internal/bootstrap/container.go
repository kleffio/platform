package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"

	// Domain module HTTP handlers
	identityhttp "github.com/kleff/platform/internal/core/identity/adapters/http"
	organizationshttp "github.com/kleff/platform/internal/core/organizations/adapters/http"
	deploymentshttp "github.com/kleff/platform/internal/core/deployments/adapters/http"
	nodeshttp "github.com/kleff/platform/internal/core/nodes/adapters/http"
	billinghttp "github.com/kleff/platform/internal/core/billing/adapters/http"
	usagehttp "github.com/kleff/platform/internal/core/usage/adapters/http"
	audithttp "github.com/kleff/platform/internal/core/audit/adapters/http"
	adminhttp "github.com/kleff/platform/internal/core/admin/adapters/http"
)

// Container holds all wired-up application components.
// This is the composition root — dependencies flow in one direction, from
// infrastructure outward to the HTTP layer.
type Container struct {
	Config *Config
	Logger *slog.Logger
	DB     *sql.DB

	// HTTP handler groups per domain module
	IdentityHandler      *identityhttp.Handler
	OrganizationsHandler *organizationshttp.Handler
	DeploymentsHandler   *deploymentshttp.Handler
	NodesHandler         *nodeshttp.Handler
	BillingHandler       *billinghttp.Handler
	UsageHandler         *usagehttp.Handler
	AuditHandler         *audithttp.Handler
	AdminHandler         *adminhttp.Handler
}

// NewContainer wires up all dependencies and returns the composition root.
func NewContainer(cfg *Config, logger *slog.Logger) (*Container, error) {
	db, err := openDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &Container{
		Config: cfg,
		Logger: logger,
		DB:     db,

		IdentityHandler:      identityhttp.NewHandler(logger),
		OrganizationsHandler: organizationshttp.NewHandler(logger),
		DeploymentsHandler:   deploymentshttp.NewHandler(logger),
		NodesHandler:         nodeshttp.NewHandler(logger),
		BillingHandler:       billinghttp.NewHandler(logger),
		UsageHandler:         usagehttp.NewHandler(logger),
		AuditHandler:         audithttp.NewHandler(logger),
		AdminHandler:         adminhttp.NewHandler(logger),
	}, nil
}

func openDatabase(dsn string) (*sql.DB, error) {
	// TODO: import a PostgreSQL driver (e.g. _ "github.com/jackc/pgx/v5/stdlib")
	// and replace this stub with database.Connect from internal/shared/database.
	// Until then, the binary will start but all persistence calls will no-op.
	_ = dsn
	return nil, nil
}
