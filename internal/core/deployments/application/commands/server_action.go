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
	constructs, err := h.catalog.ListConstructs(ctx, "", blueprint.ConstructID)
	if err != nil || len(constructs) == 0 {
		return fmt.Errorf("construct not found for blueprint")
	}
	construct := constructs[0]

	portReqs := make([]queue.PortRequirement, 0, len(construct.Ports))
	for _, p := range construct.Ports {
		portReqs = append(portReqs, queue.PortRequirement{
			TargetPort: p.Container,
			Protocol:   p.Protocol,
		})
	}

	spec := queue.WorkloadSpec{
		OwnerID:          deployment.OrganizationID,
		ServerID:         deployment.ServerName,
		BlueprintID:      blueprint.ID,
		Image:            construct.Image,
		MemoryBytes:      int64(blueprint.Resources.MemoryMB) * 1024 * 1024,
		CPUMillicores:    int64(blueprint.Resources.CPUMillicores),
		PortRequirements: portReqs,
		RuntimeHints: queue.RuntimeHints{
			KubernetesStrategy: construct.RuntimeHints.KubernetesStrategy,
			ExposeUDP:          construct.RuntimeHints.ExposeUDP,
			PersistentStorage:  construct.RuntimeHints.PersistentStorage,
			StoragePath:        construct.RuntimeHints.StoragePath,
			StorageGB:          construct.RuntimeHints.StorageGB,
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
