package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/kleff/platform/internal/core/identity/domain"
	"github.com/kleff/platform/internal/core/identity/ports"
	"github.com/kleff/platform/internal/shared/ids"
)

// CreateUserCommand carries the intent to provision a new user account,
// typically triggered after a successful OIDC sign-in.
type CreateUserCommand struct {
	ExternalID  string // OIDC subject
	Email       string
	DisplayName string
	AvatarURL   string
}

// CreateUserResult is the output of the command.
type CreateUserResult struct {
	UserID string
}

// CreateUserHandler executes CreateUserCommand.
type CreateUserHandler struct {
	users ports.UserRepository
}

func NewCreateUserHandler(users ports.UserRepository) *CreateUserHandler {
	return &CreateUserHandler{users: users}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error) {
	// Idempotent — if the user already exists by external ID, return existing.
	existing, err := h.users.FindByExternalID(ctx, cmd.ExternalID)
	if err == nil {
		return &CreateUserResult{UserID: existing.ID}, nil
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:          ids.New(),
		ExternalID:  cmd.ExternalID,
		Email:       cmd.Email,
		DisplayName: cmd.DisplayName,
		AvatarURL:   cmd.AvatarURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.users.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}

	return &CreateUserResult{UserID: user.ID}, nil
}
