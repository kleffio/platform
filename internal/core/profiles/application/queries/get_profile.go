// Package queries contains read-side application handlers for the profiles module.
package queries

import (
	"context"
	"fmt"

	"github.com/kleff/platform/internal/core/profiles/domain"
	"github.com/kleff/platform/internal/core/profiles/ports"
)

// GetProfileQuery fetches a user profile by Kratos identity ID.
type GetProfileQuery struct {
	IdentityID string
}

// GetProfileHandler executes GetProfileQuery.
type GetProfileHandler struct {
	profiles ports.ProfileRepository
}

func NewGetProfileHandler(profiles ports.ProfileRepository) *GetProfileHandler {
	return &GetProfileHandler{profiles: profiles}
}

func (h *GetProfileHandler) Handle(ctx context.Context, q GetProfileQuery) (*domain.UserProfile, error) {
	profile, err := h.profiles.FindByID(ctx, q.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("get profile %s: %w", q.IdentityID, err)
	}
	return profile, nil
}
