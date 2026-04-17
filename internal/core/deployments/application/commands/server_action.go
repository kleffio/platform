package commands

import (
	"context"
	"fmt"

	catalogports "github.com/kleffio/platform/internal/core/catalog/ports"
	"github.com/kleffio/platform/internal/core/deployments/domain"
	"github.com/kleffio/platform/internal/core/deployments/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/queue"
)

type ServerActionCommand struct {
	ServerName string
	Action     queue.JobType // JobTypeServerStop, JobTypeServerStart, JobTypeServerRestart
}

type ServerActionHandler struct {
	deployments ports.DeploymentRepository
	catalog     catalogports.CatalogRepository
	enqueuer    queue.Enqueuer
}

func NewServerActionHandler(
	deployments ports.DeploymentRepository,
	catalog catalogports.CatalogRepository,
	enqueuer queue.Enqueuer,
) *ServerActionHandler {
	return &ServerActionHandler{deployments: deployments, catalog: catalog, enqueuer: enqueuer}
}

func (h *ServerActionHandler) Handle(ctx context.Context, cmd ServerActionCommand) error {
	deployment, err := h.deployments.FindByServerID(ctx, cmd.ServerName)
	if err != nil {
		return fmt.Errorf("server not found: %w", err)
	}

	// Validate the action is allowed for the current status.
	switch cmd.Action {
	case queue.JobTypeServerStop, queue.JobTypeServerRestart:
		if deployment.Status != domain.DeploymentSucceeded {
			return fmt.Errorf("server must be running to %s (current status: %s)", actionName(cmd.Action), deployment.Status)
		}
	case queue.JobTypeServerStart:
		if deployment.Status != domain.DeploymentFailed && deployment.Status != domain.DeploymentRolledBack {
			return fmt.Errorf("server must be stopped to start (current status: %s)", deployment.Status)
		}
	case queue.JobTypeServerDelete:
		// Delete is allowed from any status.
	}

	// Rebuild the workload spec from the catalog so the daemon has everything it needs.
	blueprint, err := h.catalog.GetBlueprint(ctx, deployment.BlueprintID)
	if err != nil {
		return fmt.Errorf("blueprint not found: %w", err)
	}
	portReqs := make([]queue.PortRequirement, 0, len(blueprint.Ports))
	for _, p := range blueprint.Ports {
		portReqs = append(portReqs, queue.PortRequirement{
			TargetPort: p.Container,
			Protocol:   p.Protocol,
		})
	}

	spec := queue.WorkloadSpec{
		OwnerID:          deployment.OrganizationID,
		ServerID:         deployment.ServerName,
		BlueprintID:      blueprint.ID,
		Image:            blueprint.Image,
		MemoryBytes:      int64(blueprint.Resources.MemoryMB) * 1024 * 1024,
		CPUMillicores:    int64(blueprint.Resources.CPUMillicores),
		PortRequirements: portReqs,
		RuntimeHints: queue.RuntimeHints{
			KubernetesStrategy: blueprint.RuntimeHints.KubernetesStrategy,
			ExposeUDP:          blueprint.RuntimeHints.ExposeUDP,
			PersistentStorage:  blueprint.RuntimeHints.PersistentStorage,
			StoragePath:        blueprint.RuntimeHints.StoragePath,
			StorageGB:          blueprint.RuntimeHints.StorageGB,
		},
	}

	if err := h.enqueuer.EnqueueAction(ctx, ids.New(), cmd.Action, spec); err != nil {
		return fmt.Errorf("enqueue %s job: %w", actionName(cmd.Action), err)
	}

	return nil
}

func actionName(t queue.JobType) string {
	switch t {
	case queue.JobTypeServerStop:
		return "stop"
	case queue.JobTypeServerStart:
		return "start"
	case queue.JobTypeServerRestart:
		return "restart"
	case queue.JobTypeServerDelete:
		return "delete"
	default:
		return string(t)
	}
}
