package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/deployments/application/commands"
	"github.com/kleffio/platform/internal/core/deployments/domain"
	"github.com/kleffio/platform/internal/core/deployments/ports"
	"github.com/kleffio/platform/internal/shared/queue"
)

const basePath = "/api/v1/deployments"

// Handler groups all HTTP endpoints for the deployments module.
type Handler struct {
	create_ *commands.CreateDeploymentHandler
	action  *commands.ServerActionHandler
	repo    ports.DeploymentRepository
	secret  string // shared secret for daemon callbacks
	logger  *slog.Logger
}

func NewHandler(create *commands.CreateDeploymentHandler, action *commands.ServerActionHandler, repo ports.DeploymentRepository, secret string, logger *slog.Logger) *Handler {
	return &Handler{create_: create, action: action, repo: repo, secret: secret, logger: logger}
}

// RegisterRoutes attaches authenticated deployment routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Post(basePath, h.create)
	r.Get(basePath+"/{id}", h.get)
	r.Post(basePath+"/{id}/stop", h.stop)
	r.Post(basePath+"/{id}/start", h.start)
	r.Post(basePath+"/{id}/restart", h.restart)
	r.Delete(basePath+"/{id}", h.delete)
}

// RegisterInternalRoutes attaches daemon-facing routes (no user auth, shared secret only).
func (h *Handler) RegisterInternalRoutes(r chi.Router) {
	r.Post("/internal/deployments/{serverID}/address", h.reportAddress)
	r.Post("/internal/deployments/{serverID}/status", h.reportStatus)
}

// --- request/response types ---

type createRequest struct {
	BlueprintID string            `json:"blueprint_id"`
	ServerName  string            `json:"server_name"`
	Config      map[string]string `json:"config"`
	Resources   *resourceOverride `json:"resources,omitempty"`
}

type resourceOverride struct {
	MemoryMB      int `json:"memory_mb"`
	CPUMillicores int `json:"cpu_millicores"`
}

type deploymentResponse struct {
	ID         string `json:"id"`
	ServerName string `json:"server_name"`
	Status     string `json:"status"`
	Address    string `json:"address"`
	CreatedAt  string `json:"created_at"`
}

func toResponse(d *domain.Deployment) deploymentResponse {
	return deploymentResponse{
		ID:         d.ID,
		ServerName: d.ServerName,
		Status:     string(d.Status),
		Address:    d.Address,
		CreatedAt:  d.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- handlers ---

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	orgID := r.Header.Get("X-Org-Id")
	if orgID == "" {
		orgID = "local-org"
	}
	userID := r.Header.Get("X-User-Id")
	if userID == "" {
		userID = "local-user"
	}

	cmd := commands.CreateDeploymentCommand{
		OrganizationID: orgID,
		BlueprintID:    req.BlueprintID,
		ServerName:     req.ServerName,
		Config:         req.Config,
		InitiatedBy:    userID,
	}
	if req.Resources != nil && (req.Resources.MemoryMB > 0 || req.Resources.CPUMillicores > 0) {
		cmd.Resources = &commands.ResourceOverride{
			MemoryMB:      req.Resources.MemoryMB,
			CPUMillicores: req.Resources.CPUMillicores,
		}
	}

	result, err := h.create_.Handle(r.Context(), cmd)
	if err != nil {
		h.logger.Error("create deployment failed", "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"deployment_id": result.DeploymentID})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	orgID := r.Header.Get("X-Org-Id")
	if orgID == "" {
		orgID = "local-org"
	}

	deployments, _, err := h.repo.ListByOrganization(r.Context(), orgID, 1, 100)
	if err != nil {
		h.logger.Error("list deployments failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list deployments")
		return
	}

	resp := make([]deploymentResponse, 0, len(deployments))
	for _, d := range deployments {
		resp = append(resp, toResponse(d))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	d, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	writeJSON(w, http.StatusOK, toResponse(d))
}

// reportAddress is called by the daemon after successful provisioning.
// It expects: Authorization: Bearer <shared-secret>
// Body: { "address": "host:port" }
func (h *Handler) reportAddress(w http.ResponseWriter, r *http.Request) {
	// Validate shared secret.
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	if h.secret != "" && token != h.secret {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	serverID := chi.URLParam(r, "serverID")

	var body struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Address == "" {
		writeError(w, http.StatusBadRequest, "address is required")
		return
	}

	if err := h.repo.UpdateAddress(r.Context(), serverID, body.Address); err != nil {
		h.logger.Error("update address failed", "server_id", serverID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update address")
		return
	}

	h.logger.Info("server address reported", "server_id", serverID, "address", body.Address)
	w.WriteHeader(http.StatusNoContent)
}

// reportStatus is called by the daemon after stop/start/restart completes.
func (h *Handler) reportStatus(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	token := strings.TrimPrefix(auth, "Bearer ")
	if h.secret != "" && token != h.secret {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	serverID := chi.URLParam(r, "serverID")

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeError(w, http.StatusBadRequest, "status is required")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), serverID, body.Status); err != nil {
		h.logger.Error("update status failed", "server_id", serverID, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update status")
		return
	}

	h.logger.Info("server status reported", "server_id", serverID, "status", body.Status)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stop(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerStop)
}

func (h *Handler) start(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerStart)
}

func (h *Handler) restart(w http.ResponseWriter, r *http.Request) {
	h.runAction(w, r, queue.JobTypeServerRestart)
}

func (h *Handler) runAction(w http.ResponseWriter, r *http.Request, action queue.JobType) {
	id := chi.URLParam(r, "id")
	d, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	if err := h.action.Handle(r.Context(), commands.ServerActionCommand{
		ServerName: d.ServerName,
		Action:     action,
	}); err != nil {
		h.logger.Error("server action failed", "action", action, "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Pre-update the deployment status so the next poll reflects the transition
	// immediately rather than waiting for the daemon to call back.
	var transitioning string
	switch action {
	case queue.JobTypeServerStop:
		transitioning = "rolled_back"
	case queue.JobTypeServerStart:
		transitioning = "in_progress"
	case queue.JobTypeServerRestart:
		transitioning = "restarting"
	}
	if transitioning != "" {
		if err := h.repo.UpdateStatus(r.Context(), d.GameServerID, transitioning); err != nil {
			h.logger.Warn("failed to pre-update deployment status", "action", action, "error", err)
		}
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	d, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "deployment not found")
		return
	}
	// Tell the daemon to stop and remove the container before deleting the record.
	if err := h.action.Handle(r.Context(), commands.ServerActionCommand{
		ServerName: d.ServerName,
		Action:     queue.JobTypeServerDelete,
	}); err != nil {
		h.logger.Error("enqueue delete job failed", "id", id, "error", err)
		// Don't block the delete — remove the record anyway.
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.logger.Error("delete deployment failed", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete deployment")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
