package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	orgports "github.com/kleffio/platform/internal/core/organizations/ports"
	projectports "github.com/kleffio/platform/internal/core/projects/ports"
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
	projects  projectports.ProjectRepository
	orgs      orgports.OrganizationRepository
	repo      ports.Repository
	provision *commands.ProvisionWorkloadHandler
	action    *commands.WorkloadActionHandler
	bus       *events.Bus
	logger    *slog.Logger
}

var orgSlugCleaner = regexp.MustCompile(`[^a-z0-9-]+`)

func NewHandler(projects projectports.ProjectRepository, orgs orgports.OrganizationRepository, repo ports.Repository, provision *commands.ProvisionWorkloadHandler, action *commands.WorkloadActionHandler, bus *events.Bus, logger *slog.Logger) *Handler {
	return &Handler{projects: projects, orgs: orgs, repo: repo, provision: provision, action: action, bus: bus, logger: logger}
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
	orgID := h.callerOrganizationID(r)
	if err := h.ensureProjectAccess(r.Context(), projectID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
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
		OrganizationID: orgID,
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
	orgID := h.callerOrganizationID(r)
	if err := h.ensureProjectAccess(r.Context(), projectID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
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
	orgID := h.callerOrganizationID(r)
	if orgID != "" && workload.OrganizationID != orgID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: workload does not belong to caller organization"})
		return
	}
	writeJSON(w, http.StatusOK, workload)
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	workloadID := chi.URLParam(r, "id")
	var req struct {
		Status        string  `json:"status"`
		RuntimeRef    string  `json:"runtime_ref"`
		Endpoint      string  `json:"endpoint"`
		NodeID        string  `json:"node_id"`
		ErrorMessage  string  `json:"error_message"`
		ObservedAt    string  `json:"observed_at"`
		CPUMillicores int64   `json:"cpu_millicores"`
		MemoryMB      int64   `json:"memory_mb"`
		NetworkRxMB   float64 `json:"network_rx_mb"`
		NetworkTxMB   float64 `json:"network_tx_mb"`
		DiskReadMB    float64 `json:"disk_read_mb"`
		DiskWriteMB   float64 `json:"disk_write_mb"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	status := domain.WorkloadState(req.Status)
	if !isValidWorkloadState(status) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "status must be one of: pending, running, stopped, deleted, failed"})
		return
	}

	existing, err := h.repo.FindByID(r.Context(), workloadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workload not found"})
			return
		}
		h.logger.Error("load workload before status update", "error", err, "workload_id", workloadID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to persist workload status"})
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
	if nodeID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	if existing.NodeID != "" && existing.NodeID != nodeID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: workload is bound to a different node"})
		return
	}

	update := domain.DaemonStatusUpdate{
		WorkloadID:    workloadID,
		Status:        status,
		RuntimeRef:    req.RuntimeRef,
		Endpoint:      req.Endpoint,
		NodeID:        nodeID,
		ErrorMessage:  req.ErrorMessage,
		ObservedAt:    observedAt,
		CPUMillicores: req.CPUMillicores,
		MemoryMB:      req.MemoryMB,
		NetworkRxMB:   req.NetworkRxMB,
		NetworkTxMB:   req.NetworkTxMB,
		DiskReadMB:    req.DiskReadMB,
		DiskWriteMB:   req.DiskWriteMB,
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
	orgID := h.callerOrganizationID(r)
	if err := h.ensureProjectAccess(r.Context(), projectID, orgID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	initiatedBy := ""
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok {
		initiatedBy = claims.Subject
	}

	err := h.action.Handle(r.Context(), commands.WorkloadActionCommand{
		OrganizationID: orgID,
		ProjectID:      projectID,
		WorkloadID:     workloadID,
		Action:         action,
		InitiatedBy:    initiatedBy,
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

// callerOrganizationID resolves the active org from X-Organization-ID header,
// verifying membership. Falls back to the personal org derived from JWT sub.
func (h *Handler) callerOrganizationID(r *http.Request) string {
	headerOrgID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))
	if headerOrgID == "" {
		headerOrgID = strings.TrimSpace(r.URL.Query().Get("organization_id"))
	}

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok || claims.Subject == "" {
		return headerOrgID
	}

	if headerOrgID != "" {
		// Verify membership when an explicit org is requested.
		if h.orgs != nil {
			if _, err := h.orgs.GetMember(r.Context(), headerOrgID, claims.Subject); err != nil {
				return "" // not a member — access denied at ensureProjectAccess
			}
		}
		return headerOrgID
	}

	// Personal org fallback.
	return "org-" + normalizeOrgSlug(claims.Subject)
}

func normalizeOrgSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = orgSlugCleaner.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	if s == "" {
		return "default"
	}
	return s
}

func isValidWorkloadState(state domain.WorkloadState) bool {
	switch state {
	case domain.WorkloadPending, domain.WorkloadRunning, domain.WorkloadStopped, domain.WorkloadDeleted, domain.WorkloadFailed:
		return true
	default:
		return false
	}
}

func (h *Handler) ensureProjectAccess(ctx context.Context, projectID, organizationID string) error {
	if organizationID == "" || h.projects == nil {
		return nil
	}
	project, err := h.projects.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project.OrganizationID != organizationID {
		return fmt.Errorf("forbidden: project does not belong to caller organization")
	}
	return nil
}
