// Package commands contains write-side application handlers for the profiles module.
package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/kleff/platform/internal/core/profiles/domain"
	"github.com/kleff/platform/internal/core/profiles/ports"
)

// UpsertProfileCommand is triggered on the user's first authenticated request
// (Lazy Creation strategy). If a profile already exists for the given identity
// ID it is returned unchanged; otherwise a new default profile is created and
// persisted.
//
// Kratos integration note:
//   IdentityID must be the value of the `sub` claim on the OIDC access token,
//   which equals the Kratos identity.id UUID. The auth middleware populates
//   Claims.Subject with exactly this value.
type UpsertProfileCommand struct {
	IdentityID string // Kratos identity.id / OIDC sub
}

// UpsertProfileHandler executes UpsertProfileCommand.
type UpsertProfileHandler struct {
	profiles ports.ProfileRepository
}

func NewUpsertProfileHandler(profiles ports.ProfileRepository) *UpsertProfileHandler {
	return &UpsertProfileHandler{profiles: profiles}
}

func (h *UpsertProfileHandler) Handle(ctx context.Context, cmd UpsertProfileCommand) (*domain.UserProfile, error) {
	// Fast path — profile already exists.
	existing, err := h.profiles.FindByID(ctx, cmd.IdentityID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ports.ErrNotFound) {
		return nil, fmt.Errorf("find profile: %w", err)
	}

	// Slow path — first time this identity has hit our API. Bootstrap a default profile.
	profile := domain.NewProfile(cmd.IdentityID)
	if err := h.profiles.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}
