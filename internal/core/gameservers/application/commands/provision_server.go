package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/kleff/platform/internal/core/catalog/ports"
	"github.com/kleff/platform/internal/core/gameservers/domain"
	gsports "github.com/kleff/platform/internal/core/gameservers/ports"
	"github.com/kleff/platform/internal/shared/ids"
)

// ProvisionServerCommand holds everything needed to create and provision a
// new game server instance from a blueprint.
type ProvisionServerCommand struct {
	OrganizationID string
	OwnerID        string
	BlueprintID    string
	Name           string
	EnvOverrides   map[string]string // user-supplied overrides on top of blueprint defaults
	MemoryBytes    int64
	CPUMillicores  int64
}

// ProvisionServerResult carries the new server ID.
type ProvisionServerResult struct {
	ServerID string
}

// ProvisionServerHandler creates a GameServer record and enqueues a provision
// job for the daemon to pick up and execute via the k8s runtime.
type ProvisionServerHandler struct {
	blueprints  ports.BlueprintRepository
	gameservers gsports.GameServerRepository
	queue       gsports.GameServerQueue
}

func NewProvisionServerHandler(
	blueprints ports.BlueprintRepository,
	gameservers gsports.GameServerRepository,
	queue gsports.GameServerQueue,
) *ProvisionServerHandler {
	return &ProvisionServerHandler{blueprints: blueprints, gameservers: gameservers, queue: queue}
}

func (h *ProvisionServerHandler) Handle(ctx context.Context, cmd ProvisionServerCommand) (*ProvisionServerResult, error) {
	bp, err := h.blueprints.GetBlueprint(ctx, cmd.BlueprintID)
	if err != nil {
		return nil, fmt.Errorf("blueprint %q not found: %w", cmd.BlueprintID, err)
	}

	// Merge env: blueprint defaults → user overrides.
	env := make(map[string]string, len(bp.EnvDefaults))
	for k, v := range bp.EnvDefaults {
		env[k] = v
	}
	for k, v := range cmd.EnvOverrides {
		env[k] = v
	}

	// Build port requirements from the blueprint.
	var portReqs []gsports.PortRequirement
	for _, p := range bp.Ports {
		portReqs = append(portReqs, gsports.PortRequirement{
			TargetPort: p.ContainerPort,
			Protocol:   p.Protocol,
		})
	}

	now := time.Now().UTC()
	gs := &domain.GameServer{
		ID:             ids.New(),
		OrganizationID: cmd.OrganizationID,
		OwnerID:        cmd.OwnerID,
		BlueprintID:    cmd.BlueprintID,
		Name:           cmd.Name,
		Status:         domain.StatusProvisioning,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.gameservers.Save(ctx, gs); err != nil {
		return nil, fmt.Errorf("save game server: %w", err)
	}

	if err := h.queue.Publish(ctx, gsports.ServerJob{
		JobType:          gsports.JobTypeProvision,
		OwnerID:          cmd.OwnerID,
		ServerID:         gs.ID,
		BlueprintID:      cmd.BlueprintID,
		Image:            bp.Image,
		EnvOverrides:     env,
		MemoryBytes:      cmd.MemoryBytes,
		CPUMillicores:    cmd.CPUMillicores,
		PortRequirements: portReqs,
	}); err != nil {
		return nil, fmt.Errorf("enqueue provision job: %w", err)
	}

	return &ProvisionServerResult{ServerID: gs.ID}, nil
}
