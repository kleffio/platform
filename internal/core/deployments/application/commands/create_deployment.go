package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/kleffio/platform/internal/core/deployments/domain"
	"github.com/kleffio/platform/internal/core/deployments/ports"
	"github.com/kleffio/platform/internal/shared/ids"
)

// CreateDeploymentCommand represents user intent to deploy a specific version
// of a game server. The control plane records the intent and emits a job
// for the daemon to act upon.
type CreateDeploymentCommand struct {
	OrganizationID string
	GameServerID   string
	Version        string
	InitiatedBy    string // user ID
}

// CreateDeploymentResult carries the new deployment ID.
type CreateDeploymentResult struct {
	DeploymentID string
}

// CreateDeploymentHandler executes CreateDeploymentCommand.
type CreateDeploymentHandler struct {
	deployments ports.DeploymentRepository
	// TODO: inject event bus to publish DeploymentRequested domain event
}

func NewCreateDeploymentHandler(deployments ports.DeploymentRepository) *CreateDeploymentHandler {
	return &CreateDeploymentHandler{deployments: deployments}
}

func (h *CreateDeploymentHandler) Handle(ctx context.Context, cmd CreateDeploymentCommand) (*CreateDeploymentResult, error) {
	if cmd.Version == "" {
		return nil, fmt.Errorf("version is required")
	}

	now := time.Now().UTC()
	d := &domain.Deployment{
		ID:             ids.New(),
		OrganizationID: cmd.OrganizationID,
		GameServerID:   cmd.GameServerID,
		Version:        cmd.Version,
		Status:         domain.DeploymentPending,
		InitiatedBy:    cmd.InitiatedBy,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.deployments.Save(ctx, d); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}

	// TODO: publish DeploymentRequested event so daemon workers can pick it up.

	return &CreateDeploymentResult{DeploymentID: d.ID}, nil
}
