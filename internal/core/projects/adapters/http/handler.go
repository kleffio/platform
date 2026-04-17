package http

import (
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
	"github.com/kleffio/platform/internal/core/projects/domain"
	"github.com/kleffio/platform/internal/core/projects/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/middleware"
)

const basePath = "/api/v1/projects"

var slugCleaner = regexp.MustCompile(`[^a-z0-9-]+`)

type Handler struct {
	repo   ports.ProjectRepository
	logger *slog.Logger
}

func NewHandler(repo ports.ProjectRepository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Project CRUD
	r.Get(basePath, h.list)
	r.Post(basePath, h.create)
	r.Get(basePath+"/{id}", h.get)

	// Connections (workload edges)
	r.Get(basePath+"/{id}/connections", h.listConnections)
	r.Post(basePath+"/{id}/connections", h.createConnection)
	r.Delete(basePath+"/{id}/connections/{connID}", h.deleteConnection)

	// Graph node positions
	r.Get(basePath+"/{id}/graph-nodes", h.listGraphNodes)
	r.Put(basePath+"/{id}/graph-nodes/{workloadID}", h.upsertGraphNode)
}

// ── Project CRUD ─────────────────────────────────────────────────────────────

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	orgID, err := organizationIDFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := h.repo.EnsureOrganization(r.Context(), orgID, "Organization "+orgID); err != nil {
		h.logger.Error("ensure organization", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ensure organization"})
		return
	}

	projects, err := h.repo.ListByOrganization(r.Context(), orgID)
	if err != nil {
		h.logger.Error("list projects", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list projects"})
		return
	}
	if len(projects) == 0 {
		now := time.Now().UTC()

		slug := "default"
		name := "Default"
		// Resolve any slug conflict (rare but safe).
		if _, err := h.repo.FindBySlug(r.Context(), orgID, slug); err == nil {
			slug = fmt.Sprintf("%s-%s", slug, ids.New()[:6])
		}

		defaultProject := &domain.Project{
			ID:             ids.New(),
			OrganizationID: orgID,
			Slug:           slug,
			Name:           name,
			IsDefault:      true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := h.repo.Save(r.Context(), defaultProject); err != nil {
			h.logger.Error("create default project", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create default project"})
			return
		}
		projects = []*domain.Project{defaultProject}
	}

	writeJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
		Slug           string `json:"slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	orgID, err := organizationIDFromRequest(r)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if req.OrganizationID != "" && req.OrganizationID != orgID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: organization mismatch"})
		return
	}
	if err := h.repo.EnsureOrganization(r.Context(), orgID, "Organization "+orgID); err != nil {
		h.logger.Error("ensure organization", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to ensure organization"})
		return
	}

	slug := normalizeSlug(req.Slug)
	if slug == "" {
		slug = normalizeSlug(req.Name)
	}
	if slug == "" {
		slug = "project"
	}
	if _, err := h.repo.FindBySlug(r.Context(), orgID, slug); err == nil {
		slug = fmt.Sprintf("%s-%s", slug, ids.New()[:6])
	}

	now := time.Now().UTC()
	project := &domain.Project{
		ID:             ids.New(),
		OrganizationID: orgID,
		Slug:           slug,
		Name:           req.Name,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.repo.Save(r.Context(), project); err != nil {
		h.logger.Error("save project", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create project"})
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	project, err := h.authorizedProject(r, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		if strings.Contains(strings.ToLower(err.Error()), "forbidden") {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
		return
	}
	writeJSON(w, http.StatusOK, project)
}

// ── Connections ───────────────────────────────────────────────────────────────

func (h *Handler) listConnections(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	conns, err := h.repo.ListConnections(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list connections", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list connections"})
		return
	}
	if conns == nil {
		conns = []*domain.Connection{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"connections": conns})
}

func (h *Handler) createConnection(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	var req struct {
		SourceWorkloadID string `json:"source_workload_id"`
		TargetWorkloadID string `json:"target_workload_id"`
		Kind             string `json:"kind"`
		Label            string `json:"label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if req.SourceWorkloadID == "" || req.TargetWorkloadID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "source_workload_id and target_workload_id are required"})
		return
	}
	if req.Kind == "" {
		req.Kind = "network"
	}
	if req.Kind != "network" && req.Kind != "dependency" && req.Kind != "traffic" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "kind must be network, dependency, or traffic"})
		return
	}

	conn := &domain.Connection{
		ID:               ids.New(),
		ProjectID:        projectID,
		SourceWorkloadID: req.SourceWorkloadID,
		TargetWorkloadID: req.TargetWorkloadID,
		Kind:             req.Kind,
		Label:            req.Label,
		CreatedAt:        time.Now().UTC(),
	}
	if err := h.repo.CreateConnection(r.Context(), conn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "source or target workload was not found in this project"})
			return
		}
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "project_connections_unique") || strings.Contains(lower, "duplicate key") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "connection already exists"})
			return
		}
		h.logger.Error("create connection", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create connection"})
		return
	}
	writeJSON(w, http.StatusCreated, conn)
}

func (h *Handler) deleteConnection(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	connID := chi.URLParam(r, "connID")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	// Verify it belongs to this project.
	conn, err := h.repo.FindConnection(r.Context(), connID)
	if err != nil || conn.ProjectID != projectID {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "connection not found"})
		return
	}

	if err := h.repo.DeleteConnection(r.Context(), connID); err != nil {
		h.logger.Error("delete connection", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete connection"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Graph nodes ───────────────────────────────────────────────────────────────

func (h *Handler) listGraphNodes(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	nodes, err := h.repo.ListGraphNodes(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list graph nodes", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list graph nodes"})
		return
	}
	if nodes == nil {
		nodes = []*domain.GraphNode{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"graph_nodes": nodes})
}

func (h *Handler) upsertGraphNode(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	workloadID := chi.URLParam(r, "workloadID")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "project not found"})
			return
		}
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	var req struct {
		PositionX float64 `json:"position_x"`
		PositionY float64 `json:"position_y"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}

	node := &domain.GraphNode{
		ID:         ids.New(),
		ProjectID:  projectID,
		WorkloadID: workloadID,
		PositionX:  req.PositionX,
		PositionY:  req.PositionY,
	}
	if err := h.repo.UpsertGraphNode(r.Context(), node); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "workload was not found in this project"})
			return
		}
		h.logger.Error("upsert graph node", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save position"})
		return
	}
	writeJSON(w, http.StatusOK, node)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func organizationIDFromRequest(r *http.Request) (string, error) {
	queryOrgID := strings.TrimSpace(r.URL.Query().Get("organization_id"))
	headerOrgID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))

	if queryOrgID != "" && headerOrgID != "" && queryOrgID != headerOrgID {
		return "", fmt.Errorf("forbidden: conflicting organization context")
	}

	requestedOrgID := queryOrgID
	if requestedOrgID == "" {
		requestedOrgID = headerOrgID
	}

	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok {
		if claims.Subject != "" {
			callerOrgID := "org-" + normalizeSlug(claims.Subject)
			if requestedOrgID != "" && requestedOrgID != callerOrgID {
				return "", fmt.Errorf("forbidden: organization mismatch")
			}
			return callerOrgID, nil
		}
	}

	if requestedOrgID != "" {
		return requestedOrgID, nil
	}
	return "org-default", nil
}

func (h *Handler) authorizedProject(r *http.Request, projectID string) (*domain.Project, error) {
	project, err := h.repo.FindByID(r.Context(), projectID)
	if err != nil {
		return nil, err
	}

	orgID, err := organizationIDFromRequest(r)
	if err != nil {
		return nil, err
	}
	if orgID != "" && project.OrganizationID != orgID {
		return nil, fmt.Errorf("forbidden: project does not belong to caller organization")
	}
	return project, nil
}

func normalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	s = slugCleaner.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
