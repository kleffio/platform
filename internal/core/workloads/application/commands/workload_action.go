package commands

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	projectports "github.com/kleffio/platform/internal/core/projects/ports"
	"github.com/kleffio/platform/internal/core/workloads/domain"
	"github.com/kleffio/platform/internal/core/workloads/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/queue"
)

type WorkloadActionCommand struct {
	ProjectID   string
	WorkloadID  string
	Action      queue.JobType
	InitiatedBy string
}

type WorkloadActionHandler struct {
	workloads ports.Repository
	projects  projectports.ProjectRepository
	queue     queue.Publisher
	logger    *slog.Logger
}

func NewWorkloadActionHandler(workloads ports.Repository, projects projectports.ProjectRepository, queuePublisher queue.Publisher, logger *slog.Logger) *WorkloadActionHandler {
	return &WorkloadActionHandler{
		workloads: workloads,
		projects:  projects,
		queue:     queuePublisher,
		logger:    logger,
	}
}

func (h *WorkloadActionHandler) Handle(ctx context.Context, cmd WorkloadActionCommand) error {
	if cmd.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}
	if cmd.WorkloadID == "" {
		return fmt.Errorf("workload_id is required")
	}
	switch cmd.Action {
	case queue.JobTypeServerStart, queue.JobTypeServerStop, queue.JobTypeServerRestart, queue.JobTypeServerDelete:
	default:
		return fmt.Errorf("unsupported action: %s", cmd.Action)
	}

	workload, err := h.workloads.FindByID(ctx, cmd.WorkloadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("workload not found")
		}
		return fmt.Errorf("find workload: %w", err)
	}
	if workload.ProjectID != cmd.ProjectID {
		return fmt.Errorf("workload does not belong to project")
	}

	project, err := h.projects.FindByID(ctx, cmd.ProjectID)
	if err != nil {
		return fmt.Errorf("project not found: %w", err)
	}

	spec := ports.WorkloadSpec{
		OwnerID:       workload.OwnerID,
		ServerID:      workload.ID,
		BlueprintID:   workload.BlueprintID,
		ProjectID:     workload.ProjectID,
		ProjectSlug:   project.Slug,
		Image:         workload.Image,
		MemoryBytes:   0,
		CPUMillicores: 0,
	}
	job, err := queue.NewJob(cmd.Action, workload.ID, spec, 5)
	if err != nil {
		return fmt.Errorf("build action queue job: %w", err)
	}
	if err := h.queue.Enqueue(ctx, job); err != nil {
		return fmt.Errorf("enqueue %s workload action: %w", cmd.Action, err)
	}

	if err := h.workloads.UpdateState(ctx, workload.ID, domain.WorkloadPending, ""); err != nil && !errors.Is(err, sql.ErrNoRows) {
		h.logger.Warn("pre-update workload state failed", "workload_id", workload.ID, "error", err)
	}

	if err := h.workloads.SaveDeployment(ctx, &ports.DeploymentRecord{
		ID:             ids.New(),
		OrganizationID: workload.OrganizationID,
		ProjectID:      workload.ProjectID,
		WorkloadID:     workload.ID,
		Action:         string(cmd.Action),
		Status:         string(domain.WorkloadPending),
		InitiatedBy:    cmd.InitiatedBy,
	}); err != nil {
		h.logger.Warn("save workload deployment action failed", "workload_id", workload.ID, "action", cmd.Action, "error", err)
	}

	return nil
}
