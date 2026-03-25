package commands

import (
	"context"

	"github.com/kleff/platform/internal/core/identity/ports"
)

// ── Login ─────────────────────────────────────────────────────────────────────

// LoginCommand carries credentials for a headless sign-in.
type LoginCommand struct {
	Username string
	Password string
}

// LoginResult contains the token set returned by the identity provider.
type LoginResult struct {
	Token *ports.Token
}

// LoginHandler executes LoginCommand against the configured IdentityProvider.
type LoginHandler struct {
	idp ports.IdentityProvider
}

func NewLoginHandler(idp ports.IdentityProvider) *LoginHandler {
	return &LoginHandler{idp: idp}
}

func (h *LoginHandler) Handle(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	tok, err := h.idp.Login(ctx, cmd.Username, cmd.Password)
	if err != nil {
		return nil, err
	}
	return &LoginResult{Token: tok}, nil
}

// ── Register ──────────────────────────────────────────────────────────────────

// RegisterCommand carries the data required to create a new user account.
type RegisterCommand struct {
	Email     string
	Username  string
	Password  string
	FirstName string
	LastName  string
}

// RegisterResult contains the provider-assigned user ID of the new account.
type RegisterResult struct {
	UserID string
}

// RegisterHandler executes RegisterCommand against the configured IdentityProvider.
type RegisterHandler struct {
	idp ports.IdentityProvider
}

func NewRegisterHandler(idp ports.IdentityProvider) *RegisterHandler {
	return &RegisterHandler{idp: idp}
}

func (h *RegisterHandler) Handle(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	userID, err := h.idp.Register(ctx, ports.RegisterRequest{
		Email:     cmd.Email,
		Username:  cmd.Username,
		Password:  cmd.Password,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
	})
	if err != nil {
		return nil, err
	}
	return &RegisterResult{UserID: userID}, nil
}
