package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate runs all pending SQL migration files in order.
// Applied migrations are tracked in the schema_migrations table so each file
// is executed exactly once, even across restarts.
func Migrate(db *sql.DB, logger *slog.Logger) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("migrate: ensure migrations table: %w", err)
	}

	files, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("migrate: list migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		name := migrationName(file)

		applied, err := isMigrationApplied(db, name)
		if err != nil {
			return fmt.Errorf("migrate: check %s: %w", name, err)
		}
		if applied {
			continue
		}

		contents, err := migrationsFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}

		if err := runMigration(db, name, string(contents)); err != nil {
			return fmt.Errorf("migrate: run %s: %w", name, err)
		}

		logger.Info("migration applied", "name", name)
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name       TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func isMigrationApplied(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(context.Background(),
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = $1)`, name,
	).Scan(&exists)
	return exists, err
}

func runMigration(db *sql.DB, name, sql string) error {
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(context.Background(), sql); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(),
		`INSERT INTO schema_migrations (name) VALUES ($1)`, name,
	); err != nil {
		return err
	}

	return tx.Commit()
}

// migrationName strips the directory prefix and .sql suffix.
func migrationName(path string) string {
	name := path
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".sql")
	return name
}
