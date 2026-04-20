package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	usagedomain "github.com/kleffio/platform/internal/core/usage/domain"
	usageports "github.com/kleffio/platform/internal/core/usage/ports"
)

const basePath = "/api/v1/usage"

type Handler struct {
	repo   usageports.UsageRepository
	logger *slog.Logger
}

func NewHandler(repo usageports.UsageRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath+"/metrics", h.getMetrics)
}

// getMetrics returns the latest per-workload metrics snapshot for a project.
// Query param: project_id (required)
func (h *Handler) getMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "project_id is required"})
		return
	}

	metrics, err := h.repo.ListLatestByProject(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list metrics by project", "error", err, "project_id", projectID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch metrics"})
		return
	}

	if metrics == nil {
		metrics = []*usagedomain.WorkloadMetrics{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"workloads": metrics})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
