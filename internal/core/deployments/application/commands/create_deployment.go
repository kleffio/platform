package commands

import (
	"context"
	"fmt"
	"regexp"
	"time"

	catalogports "github.com/kleffio/platform/internal/core/catalog/ports"
	"github.com/kleffio/platform/internal/core/deployments/domain"
	"github.com/kleffio/platform/internal/core/deployments/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/queue"
)

// ResourceOverride lets the caller replace the blueprint's default resource allocation.
type ResourceOverride struct {
	MemoryMB      int
	CPUMillicores int
}

// CreateDeploymentCommand represents user intent to provision a game server
// from a blueprint. The control plane records the intent, builds a WorkloadSpec,
// and enqueues a job for the daemon to execute.
type CreateDeploymentCommand struct {
	OrganizationID string
	BlueprintID    string
	ServerName     string            // becomes the container/pod name
	Config         map[string]string // user-supplied env overrides (from form)
	InitiatedBy    string            // user ID
	Resources      *ResourceOverride // optional; falls back to blueprint defaults when nil
}

// CreateDeploymentResult carries the new deployment ID.
type CreateDeploymentResult struct {
	DeploymentID string
}

// CreateDeploymentHandler executes CreateDeploymentCommand.
type CreateDeploymentHandler struct {
	deployments ports.DeploymentRepository
	catalog     catalogports.CatalogRepository
	enqueuer    queue.Enqueuer
}

func NewCreateDeploymentHandler(
	deployments ports.DeploymentRepository,
	catalog catalogports.CatalogRepository,
	enqueuer queue.Enqueuer,
) *CreateDeploymentHandler {
	return &CreateDeploymentHandler{
		deployments: deployments,
		catalog:     catalog,
		enqueuer:    enqueuer,
	}
}

var validContainerName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.\-]*$`)

func (h *CreateDeploymentHandler) Handle(ctx context.Context, cmd CreateDeploymentCommand) (*CreateDeploymentResult, error) {
	if cmd.BlueprintID == "" {
		return nil, fmt.Errorf("blueprint_id is required")
	}
	if cmd.ServerName == "" {
		return nil, fmt.Errorf("server_name is required")
	}
	if !validContainerName.MatchString(cmd.ServerName) {
		return nil, fmt.Errorf("server_name %q is invalid: only letters, numbers, underscores, dots, and hyphens are allowed (no spaces)", cmd.ServerName)
	}

	// Look up blueprint and its construct.
	blueprint, err := h.catalog.GetBlueprint(ctx, cmd.BlueprintID)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found: %w", err)
	}
	deploymentID := ids.New()
	now := time.Now().UTC()

	d := &domain.Deployment{
		ID:             deploymentID,
		OrganizationID: cmd.OrganizationID,
		GameServerID:   cmd.ServerName,
		ServerName:     cmd.ServerName,
		BlueprintID:    cmd.BlueprintID,
		Version:        blueprint.Version,
		Status:         domain.DeploymentPending,
		InitiatedBy:    cmd.InitiatedBy,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.deployments.Save(ctx, d); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}

	// Resolve the container image. If the blueprint offers multiple images
	// (e.g. different Java versions), the frontend sends the chosen label via
	// cmd.Config["IMAGE"]. We look that label up in blueprint.Images and use
	// the resolved URL, then strip "IMAGE" from the env so it isn't forwarded
	// to the container. If no selection is made, fall back to blueprint.Image.
	image := blueprint.Image
	if len(blueprint.Images) > 0 {
		if label, ok := cmd.Config["IMAGE"]; ok && label != "" {
			if resolved, ok := blueprint.Images[label]; ok {
				image = resolved
			}
		}
	}

	// Build env: blueprint base env + user config overrides.
	env := make(map[string]string, len(blueprint.Env)+len(cmd.Config))
	for k, v := range blueprint.Env {
		env[k] = v
	}
	for k, v := range cmd.Config {
		if k == "IMAGE" {
			continue // not a container env var
		}
		if v != "" {
			env[k] = v
		}
	}
	if blueprint.StartupScript != "" {
		env["STARTUP_SCRIPT"] = blueprint.StartupScript
	}

	portReqs := make([]queue.PortRequirement, 0, len(blueprint.Ports))
	for _, p := range blueprint.Ports {
		portReqs = append(portReqs, queue.PortRequirement{
			TargetPort: p.Container,
			Protocol:   p.Protocol,
		})
	}

	memoryMB := blueprint.Resources.MemoryMB
	cpuMillicores := blueprint.Resources.CPUMillicores
	if cmd.Resources != nil {
		if cmd.Resources.MemoryMB > 0 {
			memoryMB = cmd.Resources.MemoryMB
		}
		if cmd.Resources.CPUMillicores > 0 {
			cpuMillicores = cmd.Resources.CPUMillicores
		}
	}

	spec := queue.WorkloadSpec{
		OwnerID:          cmd.OrganizationID,
		ServerID:         cmd.ServerName,
		BlueprintID:      cmd.BlueprintID,
		Image:            image,
		EnvOverrides:     env,
		MemoryBytes:      int64(memoryMB) * 1024 * 1024,
		CPUMillicores:    int64(cpuMillicores),
		PortRequirements: portReqs,
		RuntimeHints: queue.RuntimeHints{
			KubernetesStrategy: blueprint.RuntimeHints.KubernetesStrategy,
			ExposeUDP:          blueprint.RuntimeHints.ExposeUDP,
			PersistentStorage:  blueprint.RuntimeHints.PersistentStorage,
			StoragePath:        blueprint.RuntimeHints.StoragePath,
			StorageGB:          blueprint.RuntimeHints.StorageGB,
		},
	}

	if err := h.enqueuer.Enqueue(ctx, deploymentID, spec); err != nil {
		return nil, fmt.Errorf("enqueue provision job: %w", err)
	}

	return &CreateDeploymentResult{DeploymentID: deploymentID}, nil
}
