package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	// Domain module HTTP handlers
	adminhttp "github.com/kleff/platform/internal/core/admin/adapters/http"
	audithttp "github.com/kleff/platform/internal/core/audit/adapters/http"
	billinghttp "github.com/kleff/platform/internal/core/billing/adapters/http"
	cataloghttp "github.com/kleff/platform/internal/core/catalog/adapters/http"
	"github.com/kleff/platform/internal/core/catalog/adapters/persistence"
	"github.com/kleff/platform/internal/core/catalog/adapters/seed"
	deploymentshttp "github.com/kleff/platform/internal/core/deployments/adapters/http"
	gshttp "github.com/kleff/platform/internal/core/gameservers/adapters/http"
	gspersistence "github.com/kleff/platform/internal/core/gameservers/adapters/persistence"
	gsqueue "github.com/kleff/platform/internal/core/gameservers/adapters/queue"
	"github.com/kleff/platform/internal/core/gameservers/application/commands"
	identityhttp "github.com/kleff/platform/internal/core/identity/adapters/http"
	nodeshttp "github.com/kleff/platform/internal/core/nodes/adapters/http"
	organizationshttp "github.com/kleff/platform/internal/core/organizations/adapters/http"
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
	CatalogHandler       *cataloghttp.Handler
	GameServersHandler   *gshttp.Handler
}

// NewContainer wires up all dependencies and returns the composition root.
func NewContainer(cfg *Config, logger *slog.Logger) (*Container, error) {
	db, err := openDatabase(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// ── Catalog (read-only; seeded from YAML at startup) ─────────────────────
	catalogRepo := persistence.NewMemoryRepository()
	if err := seed.LoadDir(context.Background(), cfg.BlueprintsDir, catalogRepo); err != nil {
		return nil, fmt.Errorf("load blueprints from %q: %w", cfg.BlueprintsDir, err)
	}
	logger.Info("blueprints loaded", "dir", cfg.BlueprintsDir)

	// ── GameServers ───────────────────────────────────────────────────────────
	gsRepo := gspersistence.NewMemoryRepository()
	redisPublisher, err := gsqueue.NewRedisPublisher(cfg.RedisURL, cfg.RedisPassword, cfg.RedisTLS)
	if err != nil {
		return nil, fmt.Errorf("connect to Redis queue: %w", err)
	}

	provisionHandler := commands.NewProvisionServerHandler(catalogRepo, gsRepo, redisPublisher)
	stopHandler := commands.NewStopServerHandler(gsRepo, redisPublisher)

	return &Container{
		Config:       cfg,
		Logger:       logger,
		DB:           db,
		Introspector: hydra.NewIntrospector(cfg.HydraAdminURL),

		IdentityHandler:      identityhttp.NewHandler(logger),
		OrganizationsHandler: organizationshttp.NewHandler(logger),
		DeploymentsHandler:   deploymentshttp.NewHandler(logger),
		NodesHandler:         nodeshttp.NewHandler(logger),
		BillingHandler:       billinghttp.NewHandler(logger),
		UsageHandler:         usagehttp.NewHandler(logger),
		AuditHandler:         audithttp.NewHandler(logger),
		AdminHandler:         adminhttp.NewHandler(logger),
		CatalogHandler:       cataloghttp.NewHandler(catalogRepo, catalogRepo, logger),
		GameServersHandler:   gshttp.NewHandler(provisionHandler, stopHandler, gsRepo, logger),
	}, nil
}

func openDatabase(dsn string) (*sql.DB, error) {
	// TODO: import a PostgreSQL driver (e.g. _ "github.com/jackc/pgx/v5/stdlib")
	// and replace this stub with database.Connect from internal/shared/database.
	// Until then, the binary will start but all persistence calls will no-op.
	_ = dsn
	return nil, nil
}
