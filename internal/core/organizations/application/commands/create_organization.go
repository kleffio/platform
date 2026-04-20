package commands

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/kleffio/platform/internal/core/organizations/domain"
	"github.com/kleffio/platform/internal/core/organizations/ports"
	"github.com/kleffio/platform/internal/shared/ids"
)

// CreateOrganizationCommand carries the intent to create a new organization.
type CreateOrganizationCommand struct {
	Name      string
	CreatedBy string // user ID of the owner
}

// CreateOrganizationResult is returned on success.
type CreateOrganizationResult struct {
	OrganizationID string
	Slug           string
}

// CreateOrganizationHandler executes CreateOrganizationCommand.
type CreateOrganizationHandler struct {
	orgs ports.OrganizationRepository
}

func NewCreateOrganizationHandler(orgs ports.OrganizationRepository) *CreateOrganizationHandler {
	return &CreateOrganizationHandler{orgs: orgs}
}

func (h *CreateOrganizationHandler) Handle(ctx context.Context, cmd CreateOrganizationCommand) (*CreateOrganizationResult, error) {
	if strings.TrimSpace(cmd.Name) == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	slug := slugify(cmd.Name)

	now := time.Now().UTC()
	org := &domain.Organization{
		ID:        ids.New(),
		Name:      cmd.Name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.orgs.Save(ctx, org); err != nil {
		return nil, fmt.Errorf("save organization: %w", err)
	}

	// The caller becomes the first owner.
	if cmd.CreatedBy != "" {
		_ = h.orgs.AddMember(ctx, &domain.Member{
			OrgID:     org.ID,
			UserID:    cmd.CreatedBy,
			Role:      domain.RoleOwner,
			CreatedAt: now,
		})
	}

	return &CreateOrganizationResult{OrganizationID: org.ID, Slug: org.Slug}, nil
}

// slugify converts a name to a URL-safe slug.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
