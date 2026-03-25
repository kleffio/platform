package commands

import (
	"context"
	"fmt"

	"github.com/kleff/platform/internal/core/gameservers/domain"
	gsports "github.com/kleff/platform/internal/core/gameservers/ports"
)

type StopServerCommand struct {
	ServerID string
	OwnerID  string
}

type StopServerHandler struct {
	gameservers gsports.GameServerRepository
	queue       gsports.GameServerQueue
}

func NewStopServerHandler(gameservers gsports.GameServerRepository, queue gsports.GameServerQueue) *StopServerHandler {
	return &StopServerHandler{gameservers: gameservers, queue: queue}
}

func (h *StopServerHandler) Handle(ctx context.Context, cmd StopServerCommand) error {
	gs, err := h.gameservers.FindByID(ctx, cmd.ServerID)
	if err != nil {
		return fmt.Errorf("server %q not found: %w", cmd.ServerID, err)
	}

	if err := h.queue.Publish(ctx, gsports.ServerJob{
		JobType:  gsports.JobTypeStop,
		OwnerID:  cmd.OwnerID,
		ServerID: gs.ID,
	}); err != nil {
		return fmt.Errorf("enqueue stop job: %w", err)
	}

	return h.gameservers.UpdateStatus(ctx, gs.ID, domain.StatusStopped)
}
