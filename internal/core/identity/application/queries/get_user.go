package queries

import (
	"context"
	"fmt"

	"github.com/kleff/platform/internal/core/identity/domain"
	"github.com/kleff/platform/internal/core/identity/ports"
)

// GetUserQuery fetches a single user by their internal ID.
type GetUserQuery struct {
	UserID string
}

// GetUserHandler executes GetUserQuery.
type GetUserHandler struct {
	users ports.UserRepository
}

func NewGetUserHandler(users ports.UserRepository) *GetUserHandler {
	return &GetUserHandler{users: users}
}

func (h *GetUserHandler) Handle(ctx context.Context, q GetUserQuery) (*domain.User, error) {
	user, err := h.users.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user %s: %w", q.UserID, err)
	}
	return user, nil
}
