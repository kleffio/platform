// Package domain contains the UserProfile aggregate and its value objects.
// This module is intentionally separate from the identity module:
//   - identity handles authentication concerns (OIDC subjects, org membership)
//   - profiles handles what users choose to present (avatar, bio, preferences)
package domain

import "time"

// ThemePreference is the user's UI theme setting.
type ThemePreference string

const (
	ThemeSystem ThemePreference = "system"
	ThemeLight  ThemePreference = "light"
	ThemeDark   ThemePreference = "dark"
)

// IsValid returns true for the three recognised theme values.
func (t ThemePreference) IsValid() bool {
	return t == ThemeSystem || t == ThemeLight || t == ThemeDark
}

// UserProfile is the profile aggregate root.
//
// Kratos integration note:
//   The ID field MUST equal the Kratos identity.id (which is also the OIDC
//   `sub` claim on every access token). This is set once at creation and
//   never changed. It lets us skip a join when serving /api/v1/users/me —
//   we simply look up the profile by the sub claim we already have.
type UserProfile struct {
	// ID is the Kratos identity.id / OIDC subject claim.
	ID string

	// Username is an optional display handle, independent of Kratos traits.
	Username string

	// AvatarURL is the public URL (or relative path) of the uploaded avatar.
	// Empty string means no avatar has been set.
	AvatarURL string

	// Bio is a free-form biography shown on the user's profile page.
	Bio string

	// ThemePreference controls the panel UI theme for this user.
	ThemePreference ThemePreference

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProfile creates a default profile for a newly registered user.
// identityID must be the Kratos identity.id / OIDC sub claim.
func NewProfile(identityID string) *UserProfile {
	now := time.Now().UTC()
	return &UserProfile{
		ID:              identityID,
		ThemePreference: ThemeSystem,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}
