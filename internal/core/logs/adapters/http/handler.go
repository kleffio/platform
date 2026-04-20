package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/logs/domain"
	"github.com/kleffio/platform/internal/core/logs/ports"
)

type Handler struct {
	repo   ports.LogRepository
	logger *slog.Logger
}

func NewHandler(repo ports.LogRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

// RegisterInternalRoutes wires the daemon-facing ingest endpoint.
// Requires node token auth (called within the RequireNodeAuth middleware group).
func (h *Handler) RegisterInternalRoutes(r chi.Router) {
	r.Post("/api/v1/internal/workloads/{workloadID}/log-lines", h.ingestLines)
}

// RegisterRoutes wires the panel-facing read endpoint.
// Requires user JWT auth.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/v1/projects/{projectID}/workloads/{workloadID}/logs", h.getLogs)
}

// ingestLines accepts a batch of log lines from the daemon.
func (h *Handler) ingestLines(w http.ResponseWriter, r *http.Request) {
	workloadID := chi.URLParam(r, "workloadID")

	var body struct {
		ProjectID string `json:"project_id"`
		Lines     []struct {
			Ts     string `json:"ts"`
			Stream string `json:"stream"`
			Line   string `json:"line"`
		} `json:"lines"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid body"})
		return
	}

	lines := make([]*domain.LogLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		ts, err := parseRFC3339(l.Ts)
		if err != nil {
			continue
		}
		stream := l.Stream
		if stream == "" {
			stream = "stdout"
		}
		lines = append(lines, &domain.LogLine{
			WorkloadID: workloadID,
			ProjectID:  body.ProjectID,
			Ts:         ts,
			Stream:     stream,
			Line:       l.Line,
		})
	}

	if err := h.repo.SaveBatch(r.Context(), lines); err != nil {
		h.logger.Error("save log batch", "error", err, "workload_id", workloadID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save logs"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getLogs returns recent log lines for a workload.
func (h *Handler) getLogs(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "projectID")
	workloadID := chi.URLParam(r, "workloadID")
	_ = projectID // used for future access-control checks

	limit := 200
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 2000 {
			limit = n
		}
	}

	lines, err := h.repo.ListByWorkload(r.Context(), workloadID, limit)
	if err != nil {
		h.logger.Error("list log lines", "error", err, "workload_id", workloadID)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch logs"})
		return
	}
	if lines == nil {
		lines = []*domain.LogLine{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"lines": lines})
}

func parseRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
