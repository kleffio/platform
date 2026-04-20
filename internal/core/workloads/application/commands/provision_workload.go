package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	catalogports "github.com/kleffio/platform/internal/core/catalog/ports"
	projectdomain "github.com/kleffio/platform/internal/core/projects/domain"
	projectports "github.com/kleffio/platform/internal/core/projects/ports"
	"github.com/kleffio/platform/internal/core/workloads/domain"
	"github.com/kleffio/platform/internal/core/workloads/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/queue"
)

const provisionJobType = "server.provision"

type ProvisionWorkloadCommand struct {
	OrganizationID string
	ProjectID      string
	OwnerID        string
	ServerName     string
	BlueprintID    string
	Image          string
	InitiatedBy    string
	EnvOverrides   map[string]string
	MemoryBytes    int64
	CPUMillicores  int64
}

type ProvisionWorkloadResult struct {
	WorkloadID   string `json:"workload_id"`
	DeploymentID string `json:"deployment_id"`
}

type ProvisionWorkloadHandler struct {
	workloads ports.Repository
	projects  projectports.ProjectRepository
	queue     queue.Publisher
	catalog   catalogports.CatalogRepository
	logger    *slog.Logger
}

func NewProvisionWorkloadHandler(workloads ports.Repository, projects projectports.ProjectRepository, queuePublisher queue.Publisher, catalog catalogports.CatalogRepository, logger *slog.Logger) *ProvisionWorkloadHandler {
	return &ProvisionWorkloadHandler{
		workloads: workloads,
		projects:  projects,
		queue:     queuePublisher,
		catalog:   catalog,
		logger:    logger,
	}
}

var validWorkloadName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.\-]*$`)

func (h *ProvisionWorkloadHandler) Handle(ctx context.Context, cmd ProvisionWorkloadCommand) (*ProvisionWorkloadResult, error) {
	if cmd.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if cmd.BlueprintID == "" {
		return nil, fmt.Errorf("blueprint_id is required")
	}

	serverName := strings.TrimSpace(cmd.ServerName)
	if serverName == "" {
		serverName = ids.New()
	}
	if !validWorkloadName.MatchString(serverName) {
		return nil, fmt.Errorf("server_name %q is invalid: only letters, numbers, underscores, dots, and hyphens are allowed (no spaces)", serverName)
	}
	existingWorkload, findErr := h.workloads.FindByProjectAndName(ctx, cmd.ProjectID, serverName)
	if findErr == nil {
		if existingWorkload.State == domain.WorkloadDeleted {
			if cleanupErr := h.workloads.DeleteWorkload(ctx, existingWorkload.ID); cleanupErr != nil && !errors.Is(cleanupErr, sql.ErrNoRows) {
				return nil, fmt.Errorf("cleanup deleted workload %q: %w", serverName, cleanupErr)
			}
		} else {
			return nil, fmt.Errorf("server_name %q already exists in this project", serverName)
		}
	} else if !errors.Is(findErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("check existing workload: %w", findErr)
	}

	workloadID := ids.New()

	project, err := h.projects.FindByID(ctx, cmd.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}
	if cmd.OrganizationID != "" && project.OrganizationID != cmd.OrganizationID {
		return nil, fmt.Errorf("forbidden: project does not belong to caller organization")
	}
	if strings.TrimSpace(cmd.OwnerID) == "" {
		cmd.OwnerID = project.OrganizationID
	}

	blueprint, err := h.catalog.GetBlueprint(ctx, cmd.BlueprintID)
	if err != nil {
		return nil, fmt.Errorf("blueprint not found: %w", err)
	}

	now := time.Now().UTC()
	deploymentID := ids.New()

	// Resolve image: if the caller passed an IMAGE key in EnvOverrides (construct
	// selector), use it to look up the image URL in blueprint.Constructs.
	// Fall back to cmd.Image, then blueprint.Image.
	image := cmd.Image
	if image == "" {
		if len(blueprint.Constructs) > 0 {
			if label, ok := cmd.EnvOverrides["IMAGE"]; ok && label != "" {
				if resolved, ok := blueprint.Constructs[label]; ok {
					image = resolved
				}
			}
		}
		if image == "" {
			image = blueprint.Image
		}
	}

	memoryBytes := cmd.MemoryBytes
	if memoryBytes <= 0 {
		memoryBytes = int64(blueprint.Resources.MemoryMB) * 1024 * 1024
	}
	cpuMillicores := cmd.CPUMillicores
	if cpuMillicores <= 0 {
		cpuMillicores = int64(blueprint.Resources.CPUMillicores)
	}

	env := make(map[string]string, len(blueprint.Env)+len(cmd.EnvOverrides))
	for k, v := range blueprint.Env {
		env[k] = v
	}
	for k, v := range cmd.EnvOverrides {
		if k == "IMAGE" {
			continue // not a container env var
		}
		if strings.TrimSpace(v) != "" {
			env[k] = v
		}
	}
	if blueprint.StartupScript != "" {
		env["STARTUP_SCRIPT"] = blueprint.StartupScript
	}

	portRequirements := make([]ports.PortRequirement, 0, len(blueprint.Ports))
	for _, p := range blueprint.Ports {
		portRequirements = append(portRequirements, ports.PortRequirement{
			TargetPort: p.Container,
			Protocol:   p.Protocol,
		})
	}

	workload := &domain.Workload{
		ID:             workloadID,
		Name:           serverName,
		OrganizationID: project.OrganizationID,
		ProjectID:      project.ID,
		OwnerID:        cmd.OwnerID,
		BlueprintID:    cmd.BlueprintID,
		Image:          image,
		State:          domain.WorkloadPending,
		CPUMillicores:  cpuMillicores,
		MemoryBytes:    memoryBytes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.workloads.CreateWorkload(ctx, workload); err != nil {
		return nil, fmt.Errorf("create workload: %w", err)
	}

	if err := h.workloads.SaveDeployment(ctx, &ports.DeploymentRecord{
		ID:             deploymentID,
		OrganizationID: project.OrganizationID,
		ProjectID:      project.ID,
		WorkloadID:     workloadID,
		Action:         "provision",
		Status:         "pending",
		InitiatedBy:    cmd.InitiatedBy,
	}); err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}

	spec := ports.WorkloadSpec{
		OwnerID:          cmd.OwnerID,
		ServerID:         workloadID,
		BlueprintID:      cmd.BlueprintID,
		ProjectID:        project.ID,
		ProjectSlug:      project.Slug,
		Image:            image,
		BlueprintVersion: blueprint.Version,
		EnvOverrides:     env,
		MemoryBytes:      memoryBytes,
		CPUMillicores:    cpuMillicores,
		PortRequirements: portRequirements,
		RuntimeHints: ports.RuntimeHints{
			KubernetesStrategy: blueprint.RuntimeHints.KubernetesStrategy,
			ExposeUDP:          blueprint.RuntimeHints.ExposeUDP,
			PersistentStorage:  blueprint.RuntimeHints.PersistentStorage,
			StoragePath:        blueprint.RuntimeHints.StoragePath,
			StorageGB:          blueprint.RuntimeHints.StorageGB,
		},
	}

	job, err := queue.NewJob(provisionJobType, workloadID, spec, 5)
	if err != nil {
		return nil, fmt.Errorf("build queue job: %w", err)
	}
	if err := h.queue.Enqueue(ctx, job); err != nil {
		h.logger.Error("enqueue provision workload job", "error", err, "workload_id", workloadID)
		if cleanupErr := h.workloads.DeleteWorkload(ctx, workloadID); cleanupErr != nil && !errors.Is(cleanupErr, sql.ErrNoRows) {
			h.logger.Warn("cleanup workload after enqueue failure", "error", cleanupErr, "workload_id", workloadID)
		}
		return nil, fmt.Errorf("enqueue provision workload: %w", err)
	}

	if existing, listErr := h.workloads.ListByProject(ctx, project.ID); listErr != nil {
		h.logger.Warn("list workloads for connection suggestions", "error", listErr, "project_id", project.ID)
	} else {
		for _, conn := range buildSuggestedConnections(workload, existing) {
			createErr := h.projects.CreateConnection(ctx, &projectdomain.Connection{
				ID:               ids.New(),
				ProjectID:        project.ID,
				SourceWorkloadID: conn.SourceWorkloadID,
				TargetWorkloadID: conn.TargetWorkloadID,
				Kind:             conn.Kind,
				Label:            conn.Label,
				CreatedAt:        time.Now().UTC(),
			})
			if createErr != nil {
				h.logger.Warn(
					"create suggested connection",
					"error", createErr,
					"project_id", project.ID,
					"source_workload_id", conn.SourceWorkloadID,
					"target_workload_id", conn.TargetWorkloadID,
				)
			}
		}
	}

	return &ProvisionWorkloadResult{WorkloadID: workloadID, DeploymentID: deploymentID}, nil
}

type suggestedConnection struct {
	SourceWorkloadID string
	TargetWorkloadID string
	Kind             string
	Label            string
}

type workloadRole string

const (
	workloadRoleProxy      workloadRole = "proxy"
	workloadRoleApp        workloadRole = "app"
	workloadRoleWorker     workloadRole = "worker"
	workloadRoleDatabase   workloadRole = "database"
	workloadRoleCache      workloadRole = "cache"
	workloadRoleGameServer workloadRole = "game-server"
	workloadRoleOther      workloadRole = "other"
)

func inferWorkloadRole(image, blueprintID string) workloadRole {
	descriptor := strings.ToLower(strings.TrimSpace(image + " " + blueprintID))

	switch {
	case strings.Contains(descriptor, "proxy"),
		strings.Contains(descriptor, "nginx"),
		strings.Contains(descriptor, "envoy"),
		strings.Contains(descriptor, "traefik"):
		return workloadRoleProxy
	case strings.Contains(descriptor, "postgres"),
		strings.Contains(descriptor, "mysql"),
		strings.Contains(descriptor, "mariadb"),
		strings.Contains(descriptor, "mongo"):
		return workloadRoleDatabase
	case strings.Contains(descriptor, "redis"),
		strings.Contains(descriptor, "cache"),
		strings.Contains(descriptor, "memcached"):
		return workloadRoleCache
	case strings.Contains(descriptor, "worker"),
		strings.Contains(descriptor, "queue"),
		strings.Contains(descriptor, "job"):
		return workloadRoleWorker
	case strings.Contains(descriptor, "minecraft"),
		strings.Contains(descriptor, "game"):
		return workloadRoleGameServer
	case strings.Contains(descriptor, "web"),
		strings.Contains(descriptor, "frontend"),
		strings.Contains(descriptor, "api"),
		strings.Contains(descriptor, "backend"):
		return workloadRoleApp
	default:
		return workloadRoleOther
	}
}

func isRuntimeService(role workloadRole) bool {
	return role == workloadRoleApp || role == workloadRoleWorker || role == workloadRoleGameServer
}

func suggestedEdge(sourceRole, targetRole workloadRole) (kind, label string, ok bool) {
	if sourceRole == workloadRoleProxy && isRuntimeService(targetRole) {
		return "traffic", "Ingress routing", true
	}
	if isRuntimeService(sourceRole) && (targetRole == workloadRoleDatabase || targetRole == workloadRoleCache) {
		return "dependency", "Service dependency", true
	}
	return "", "", false
}

func buildSuggestedConnections(current *domain.Workload, peers []*domain.Workload) []suggestedConnection {
	if current == nil {
		return nil
	}

	currentRole := inferWorkloadRole(current.Image, current.BlueprintID)
	out := make([]suggestedConnection, 0)

	for _, peer := range peers {
		if peer == nil || peer.ID == "" || peer.ID == current.ID || peer.State == domain.WorkloadDeleted {
			continue
		}

		peerRole := inferWorkloadRole(peer.Image, peer.BlueprintID)

		if kind, label, ok := suggestedEdge(currentRole, peerRole); ok {
			out = append(out, suggestedConnection{
				SourceWorkloadID: current.ID,
				TargetWorkloadID: peer.ID,
				Kind:             kind,
				Label:            label,
			})
			continue
		}

		if kind, label, ok := suggestedEdge(peerRole, currentRole); ok {
			out = append(out, suggestedConnection{
				SourceWorkloadID: peer.ID,
				TargetWorkloadID: current.ID,
				Kind:             kind,
				Label:            label,
			})
		}
	}

	return out
}
