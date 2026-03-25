// Package persistence provides the PostgreSQL implementation of ports.ProfileRepository.
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/kleff/platform/internal/core/profiles/domain"
	"github.com/kleff/platform/internal/core/profiles/ports"
)

// PostgresProfileRepository implements ports.ProfileRepository.
type PostgresProfileRepository struct {
	db *sql.DB
}

func NewPostgresProfileRepository(db *sql.DB) *PostgresProfileRepository {
	return &PostgresProfileRepository{db: db}
}

// FindByID fetches the profile row whose id matches the given Kratos identity ID.
func (r *PostgresProfileRepository) FindByID(ctx context.Context, identityID string) (*domain.UserProfile, error) {
	if r.db == nil {
		// DB not yet wired — return not-found so the handler falls through to create.
		return nil, ports.ErrNotFound
	}

	const q = `
		SELECT id, username, avatar_url, bio, theme_preference, created_at, updated_at
		FROM user_profiles
		WHERE id = $1
	`

	var (
		p        domain.UserProfile
		username sql.NullString
		avatarURL sql.NullString
		bio      sql.NullString
	)

	err := r.db.QueryRowContext(ctx, q, identityID).Scan(
		&p.ID,
		&username,
		&avatarURL,
		&bio,
		&p.ThemePreference,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ports.ErrNotFound
		}
		return nil, fmt.Errorf("query profile: %w", err)
	}

	p.Username = username.String
	p.AvatarURL = avatarURL.String
	p.Bio = bio.String
	return &p, nil
}

// Save upserts a profile row. On conflict (same id) it updates all mutable columns.
func (r *PostgresProfileRepository) Save(ctx context.Context, profile *domain.UserProfile) error {
	if r.db == nil {
		// DB not yet wired — no-op so the binary still boots and responds.
		return nil
	}

	const q = `
		INSERT INTO user_profiles (id, username, avatar_url, bio, theme_preference, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			username         = EXCLUDED.username,
			avatar_url       = EXCLUDED.avatar_url,
			bio              = EXCLUDED.bio,
			theme_preference = EXCLUDED.theme_preference,
			updated_at       = $7
	`

	_, err := r.db.ExecContext(ctx, q,
		profile.ID,
		nullableString(profile.Username),
		nullableString(profile.AvatarURL),
		nullableString(profile.Bio),
		string(profile.ThemePreference),
		profile.CreatedAt.UTC(),
		time.Now().UTC(), // updated_at always reflects the save time
	)
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	return nil
}

// nullableString converts an empty string to a SQL NULL, and a non-empty
// string to a valid NullString. This keeps nullable columns clean in Postgres.
func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
