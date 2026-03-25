// Package ports defines the inbound/outbound interfaces for the profiles module.
package ports

import (
	"context"
	"errors"

	"github.com/kleff/platform/internal/core/profiles/domain"
)

// ErrNotFound is returned when a profile does not exist.
var ErrNotFound = errors.New("profile not found")

// ProfileRepository is the persistence contract for UserProfile aggregates.
// The concrete implementation lives in adapters/persistence.
type ProfileRepository interface {
	// FindByID returns the profile whose ID matches the given identity ID,
	// or ErrNotFound if no row exists yet.
	FindByID(ctx context.Context, identityID string) (*domain.UserProfile, error)

	// Save inserts or updates a profile (upsert semantics on the ID column).
	Save(ctx context.Context, profile *domain.UserProfile) error
}
