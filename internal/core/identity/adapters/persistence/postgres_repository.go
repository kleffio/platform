package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kleff/platform/internal/core/identity/domain"
)

// PostgresUserRepository implements ports.UserRepository using PostgreSQL.
type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

var errNotFound = errors.New("not found")

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	// TODO: implement with sqlx or pgx
	_ = id
	return nil, errNotFound
}

func (r *PostgresUserRepository) FindByExternalID(ctx context.Context, externalID string) (*domain.User, error) {
	// TODO: implement
	_ = externalID
	return nil, errNotFound
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	// TODO: implement
	_ = email
	return nil, errNotFound
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *domain.User) error {
	// TODO: implement with upsert
	return nil
}
