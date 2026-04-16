package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/nodes/application"
	"github.com/kleffio/platform/internal/core/nodes/domain"
	"github.com/kleffio/platform/internal/core/nodes/ports"
	"github.com/kleffio/platform/internal/shared/ids"
)

const basePath = "/api/v1/nodes"

// Handler groups all HTTP endpoints for the nodes module.
type Handler struct {
	repo   ports.NodeRepository
	logger *slog.Logger
}

func NewHandler(repo ports.NodeRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// RegisterPublicRoutes attaches node bootstrap routes.
func (h *Handler) RegisterPublicRoutes(r chi.Router) {
	r.Post(basePath, h.register)
}

// RegisterRoutes attaches all node routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Get(basePath+"/{id}", h.get)
	r.Post(basePath+"/{id}/drain", h.drain)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NodeID      string `json:"node_id"`
		Hostname    string `json:"hostname"`
		Region      string `json:"region"`
		IPAddress   string `json:"ip_address"`
		TotalVCPU   int    `json:"total_vcpu"`
		TotalMemGB  int    `json:"total_mem_gb"`
		TotalDiskGB int    `json:"total_disk_gb"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if req.Hostname == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "hostname is required"})
		return
	}

	nodeID := req.NodeID
	if nodeID != "" {
		existing, findErr := h.repo.FindByID(r.Context(), nodeID)
		switch {
		case findErr == nil:
			if existing.Hostname != req.Hostname {
				writeJSON(w, http.StatusConflict, map[string]string{"error": "node_id is already assigned to a different hostname"})
				return
			}
		case errors.Is(findErr, sql.ErrNoRows):
			// New explicit node ID registration; allowed by bootstrap secret.
		case findErr != nil:
			h.logger.Error("lookup node by id", "error", findErr)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register node"})
			return
		}
	} else {
		existing, findErr := h.repo.FindByHostname(r.Context(), req.Hostname)
		switch {
		case findErr == nil:
			nodeID = existing.ID
		case errors.Is(findErr, sql.ErrNoRows):
			nodeID = ids.New()
		default:
			h.logger.Error("lookup node by hostname", "error", findErr)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register node"})
			return
		}
	}
	if nodeID == "" {
		nodeID = ids.New()
	}
	token, err := application.GenerateNodeToken()
	if err != nil {
		h.logger.Error("generate node token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to generate node token"})
		return
	}
	now := time.Now().UTC()
	n := &domain.Node{
		ID:              nodeID,
		Hostname:        req.Hostname,
		Region:          req.Region,
		IPAddress:       req.IPAddress,
		Status:          domain.NodeStatusOnline,
		TokenHash:       application.HashNodeToken(token),
		TotalVCPU:       req.TotalVCPU,
		TotalMemGB:      req.TotalMemGB,
		TotalDiskGB:     req.TotalDiskGB,
		LastHeartbeatAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if n.Region == "" {
		n.Region = "local"
	}

	if err := h.repo.Save(r.Context(), n); err != nil {
		h.logger.Error("register node", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to register node"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"node_id":    nodeID,
		"node_token": token,
	})
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	nodes, err := h.repo.ListAll(r.Context())
	if err != nil {
		h.logger.Error("list nodes", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list nodes"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"nodes": nodes})
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	node, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("node %s not found", id)})
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (h *Handler) drain(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	node, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("node %s not found", id)})
		return
	}
	node.Status = domain.NodeStatusDraining
	node.UpdatedAt = time.Now().UTC()
	if err := h.repo.Save(r.Context(), node); err != nil {
		h.logger.Error("drain node", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to drain node"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "draining"})
}
