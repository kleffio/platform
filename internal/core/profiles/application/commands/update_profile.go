package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kleff/platform/internal/core/profiles/domain"
	"github.com/kleff/platform/internal/core/profiles/ports"
)

// UpdateProfileCommand carries only the mutable fields the user may change.
// Fields are pointers so the handler can distinguish "not provided" from "set to empty".
type UpdateProfileCommand struct {
	IdentityID      string // from claims — not user-supplied
	Bio             *string
	ThemePreference *domain.ThemePreference
	AvatarURL       *string // set by the avatar upload handler, not the PATCH endpoint
}

// UpdateProfileHandler executes UpdateProfileCommand.
type UpdateProfileHandler struct {
	profiles ports.ProfileRepository
}

func NewUpdateProfileHandler(profiles ports.ProfileRepository) *UpdateProfileHandler {
	return &UpdateProfileHandler{profiles: profiles}
}

func (h *UpdateProfileHandler) Handle(ctx context.Context, cmd UpdateProfileCommand) (*domain.UserProfile, error) {
	profile, err := h.profiles.FindByID(ctx, cmd.IdentityID)
	if err != nil {
		if errors.Is(err, ports.ErrNotFound) {
			return nil, fmt.Errorf("profile not found for identity %s", cmd.IdentityID)
		}
		return nil, fmt.Errorf("find profile: %w", err)
	}

	if cmd.Bio != nil {
		profile.Bio = *cmd.Bio
	}
	if cmd.ThemePreference != nil {
		if !cmd.ThemePreference.IsValid() {
			return nil, fmt.Errorf("invalid theme_preference %q: must be light, dark, or system", *cmd.ThemePreference)
		}
		profile.ThemePreference = *cmd.ThemePreference
	}
	if cmd.AvatarURL != nil {
		profile.AvatarURL = *cmd.AvatarURL
	}

	profile.UpdatedAt = time.Now().UTC()

	if err := h.profiles.Save(ctx, profile); err != nil {
		return nil, fmt.Errorf("save profile: %w", err)
	}
	return profile, nil
}
