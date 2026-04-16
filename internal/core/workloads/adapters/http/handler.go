package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/workloads/application/commands"
	"github.com/kleffio/platform/internal/core/workloads/domain"
	"github.com/kleffio/platform/internal/core/workloads/ports"
	"github.com/kleffio/platform/internal/shared/events"
	"github.com/kleffio/platform/internal/shared/middleware"
	"github.com/kleffio/platform/internal/shared/queue"
)

const (
	projectBasePath  = "/api/v1/projects/{projectID}/workloads"
	workloadBasePath = "/api/v1/workloads"
	internalBasePath = "/api/v1/internal/workloads"
)

type Handler struct {
	repo      ports.Repository
	provision *commands.ProvisionWorkloadHandler
	action    *commands.WorkloadActionHandler
	bus       *events.Bus
	logger    *slog.Logger
}

func NewHandler(repo ports.Repository, provision *commands.ProvisionWorkloadHandler, action *commands.WorkloadActionHandler, bus *events.Bus, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, provision: provision, action: action, bus: bus, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(projectBasePath, h.list)
	r.Post(projectBasePath, h.provisionWorkload)
	r.Post(projectBasePath+"/{id}/start", h.start)
	r.Post(projectBasePath+"/{id}/stop", h.stop)
	r.Post(projectBasePath+"/{id}/restart", h.restart)
	r.Delete(projectBasePath+"/{id}", h.delete)
	r.Get(workloadBasePath+"/{id}", h.get)
}

func (h *Handler) RegisterInternalRoutes(r chi.Router) {
	r.Post(internalBasePath+"/{id}/status", h.updateStatus)
}

func (h *Handler) provisionWorkload(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	var req struct {
		OrganizationID string            `json:"organization_id"`
		OwnerID        string            `json:"owner_id"`
		ServerName     string            `json:"server_name"`
		BlueprintID    string            `json:"blueprint_id"`
		Image          string            `json:"image"`
		Config         map[string]string `json:"config"`
		EnvOverrides   map[string]string `json:"env_overrides"`
		MemoryBytes    int64             `json:"memory_bytes"`
		CPUMillicores  int64             `json:"cpu_millicores"`
		Resources      *struct {
			MemoryMB      int64 `json:"memory_mb"`
			CPUMillicores int64 `json:"cpu_millicores"`
		} `json:"resources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	envOverrides := req.EnvOverrides
	if len(envOverrides) == 0 && len(req.Config) > 0 {
		envOverrides = req.Config
	}
	memoryBytes := req.MemoryBytes
	cpuMillicores := req.CPUMillicores
	if req.Resources != nil {
		if req.Resources.MemoryMB > 0 && memoryBytes <= 0 {
			memoryBytes = req.Resources.MemoryMB * 1024 * 1024
		}
		if req.Resources.CPUMillicores > 0 && cpuMillicores <= 0 {
			cpuMillicores = req.Resources.CPUMillicores
		}
	}
	initiatedBy := ""
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok {
		initiatedBy = claims.Subject
	}
	res, err := h.provision.Handle(r.Context(), commands.ProvisionWorkloadCommand{
		OrganizationID: req.OrganizationID,
		ProjectID:      projectID,
		OwnerID:        req.OwnerID,
		ServerName:     req.ServerName,
		BlueprintID:    req.BlueprintID,
		Image:          req.Image,
		InitiatedBy:    initiatedBy,
		EnvOverrides:   envOverrides,
		MemoryBytes:    memoryBytes,
		CPUMillicores:  cpuMillicores,
	})
	if err != nil {
		h.logger.Error("provision workload", "error", err, "project_id", projectID)
		status := http.StatusBadRequest
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "forbidden") {
			status = http.StatusForbidden
		} else if strings.Contains(lower, "not found") {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, res)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	workloads, err := h.repo.ListByProject(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list workloads", "error", err, "project_id", projectID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list workloads"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workloads": workloads})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	workload, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "workload not found"})
		return
	}
	writeJSON(w, http.StatusOK, workload)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	workloadID := chi.URLParam(r, "id")
	var req struct {
		Status       string `json:"status"`
		RuntimeRef   string `json:"runtime_ref"`
		Endpoint     string `json:"endpoint"`
		NodeID       string `json:"node_id"`
		ErrorMessage string `json:"error_message"`
		ObservedAt   string `json:"observed_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	status := domain.WorkloadState(req.Status)
	if status == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status is required"})
		return
	}

	observedAt := time.Now().UTC()
	if req.ObservedAt != "" {
		if parsed, err := time.Parse(time.RFC3339, req.ObservedAt); err == nil {
			observedAt = parsed.UTC()
		}
	}
	nodeID := req.NodeID
	if claims, ok := middleware.NodeClaimsFromContext(r.Context()); ok {
		nodeID = claims.NodeID
	}

	update := domain.DaemonStatusUpdate{
		WorkloadID:   workloadID,
		Status:       status,
		RuntimeRef:   req.RuntimeRef,
		Endpoint:     req.Endpoint,
		NodeID:       nodeID,
		ErrorMessage: req.ErrorMessage,
		ObservedAt:   observedAt,
	}
	if err := h.repo.UpdateFromDaemon(r.Context(), update); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workload not found"})
			return
		}
		h.logger.Error("update workload status", "error", err, "workload_id", workloadID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist workload status"})
		return
	}

	if h.bus != nil {
		_ = h.bus.Publish(r.Context(), domain.WorkloadStatusChanged{
			WorkloadID: workloadID,
			Status:     status,
			NodeID:     nodeID,
			Endpoint:   req.Endpoint,
		})
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerStart)
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerStop)
}

func (h *Handler) restart(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerRestart)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerDelete)
}

func (h *Handler) runAction(w http.ResponseWriter, r *http.Request, action queue.JobType) {
	projectID := chi.URLParam(r, "projectID")
	workloadID := chi.URLParam(r, "id")
	initiatedBy := ""
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok {
		initiatedBy = claims.Subject
	}

	err := h.action.Handle(r.Context(), commands.WorkloadActionCommand{
		ProjectID:   projectID,
		WorkloadID:  workloadID,
		Action:      action,
		InitiatedBy: initiatedBy,
	})
	if err != nil {
		status := http.StatusBadRequest
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "forbidden") {
			status = http.StatusForbidden
		} else if strings.Contains(lower, "not found") {
			status = http.StatusNotFound
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
