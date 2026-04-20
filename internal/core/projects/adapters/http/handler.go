package http

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/notifications/application"
	notificationsdomain "github.com/kleffio/platform/internal/core/notifications/domain"
	orgports "github.com/kleffio/platform/internal/core/organizations/ports"
	"github.com/kleffio/platform/internal/core/projects/domain"
	"github.com/kleffio/platform/internal/core/projects/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/middleware"
)

const basePath = "/api/v1/projects"

var slugCleaner = regexp.MustCompile(`[^a-z0-9-]+`)

type Handler struct {
	repo          ports.ProjectRepository
	orgs          orgports.OrganizationRepository
	notifications *application.Service
	logger        *slog.Logger
}

func NewHandler(repo ports.ProjectRepository, orgs orgports.OrganizationRepository, notifications *application.Service, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, orgs: orgs, notifications: notifications, logger: logger}
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

	// Members sub-resource
	r.Get(basePath+"/{id}/members", h.listMembers)
	r.Post(basePath+"/{id}/members", h.addMember)
	r.Patch(basePath+"/{id}/members/{userID}", h.updateMemberRole)
	r.Delete(basePath+"/{id}/members/{userID}", h.removeMember)

	// Invites sub-resource
	r.Get(basePath+"/{id}/invites", h.listInvites)
	r.Post(basePath+"/{id}/invites", h.createInvite)
	r.Delete(basePath+"/{id}/invites/{inviteID}", h.revokeInvite)

	// Public invite resolution + accept
	r.Get("/api/v1/project-invites/{token}", h.resolveInvite)
	r.Post("/api/v1/project-invites/{token}/accept", h.acceptInvite)
}

// ── Project CRUD ─────────────────────────────────────────────────────────────

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	orgID, err := h.resolveOrganizationID(r)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}

	projects, err := h.repo.ListByOrganization(r.Context(), orgID)
	if err != nil {
		h.logger.Error("list projects", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list projects"})
		return
	}

	// Also include projects in other orgs where the user is an explicit member.
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
		memberProjects, _ := h.repo.ListByMember(r.Context(), claims.Subject)
		seen := make(map[string]struct{}, len(projects))
		for _, p := range projects {
			seen[p.ID] = struct{}{}
		}
		for _, p := range memberProjects {
			if _, exists := seen[p.ID]; !exists {
				projects = append(projects, p)
			}
		}
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
		if claims, ok := middleware.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
			_ = h.repo.AddMember(r.Context(), &domain.ProjectMember{
				ProjectID:   defaultProject.ID,
				UserID:      claims.Subject,
				Email:       claims.Email,
				DisplayName: claims.Username,
				Role:        domain.RoleOwner,
				CreatedAt:   now,
			})
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
	orgID, err := h.resolveOrganizationID(r)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if req.OrganizationID != "" && req.OrganizationID != orgID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: organization mismatch"})
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
	claims, _ := middleware.ClaimsFromContext(r.Context())
	_ = h.repo.AddMember(r.Context(), &domain.ProjectMember{
		ProjectID:   project.ID,
		UserID:      claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.Username,
		Role:        domain.RoleOwner,
		CreatedAt:   now,
	})

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

// ── Members ───────────────────────────────────────────────────────────────────

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProject(r, projectID); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	members, err := h.repo.ListMembers(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list project members", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list members"})
		return
	}
	if members == nil {
		members = []*domain.ProjectMember{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (h *Handler) addMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	var req struct {
		UserID      string `json:"user_id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id is required"})
		return
	}
	if domain.RoleRank(req.Role) < 0 {
		req.Role = domain.RoleDeveloper
	}
	// Only owners can add a member with owner role.
	if req.Role == domain.RoleOwner {
		if _, err := h.authorizedProjectRole(r, projectID, domain.RoleOwner); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only an owner can add another owner"})
			return
		}
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())
	member := &domain.ProjectMember{
		ProjectID:   projectID,
		UserID:      req.UserID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        req.Role,
		InvitedBy:   claims.Subject,
		CreatedAt:   time.Now().UTC(),
	}
	if err := h.repo.AddMember(r.Context(), member); err != nil {
		h.logger.Error("add project member", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to add member"})
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *Handler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userID")
	if _, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || domain.RoleRank(req.Role) < 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "valid role is required"})
		return
	}
	// Granting owner, or modifying an existing owner, requires owner.
	target, _ := h.repo.GetMember(r.Context(), projectID, userID)
	if req.Role == domain.RoleOwner || (target != nil && target.Role == domain.RoleOwner) {
		if _, err := h.authorizedProjectRole(r, projectID, domain.RoleOwner); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only an owner can grant or change the owner role"})
			return
		}
	}
	if err := h.repo.UpdateMemberRole(r.Context(), projectID, userID, req.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "member not found"})
			return
		}
		h.logger.Error("update member role", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update role"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userID")
	if _, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	// Removing an owner requires owner.
	target, _ := h.repo.GetMember(r.Context(), projectID, userID)
	if target != nil && target.Role == domain.RoleOwner {
		if _, err := h.authorizedProjectRole(r, projectID, domain.RoleOwner); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only an owner can remove another owner"})
			return
		}
	}
	if err := h.repo.RemoveMember(r.Context(), projectID, userID); err != nil {
		h.logger.Error("remove member", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to remove member"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Invites ───────────────────────────────────────────────────────────────────

func (h *Handler) listInvites(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	if _, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	invites, err := h.repo.ListInvites(r.Context(), projectID)
	if err != nil {
		h.logger.Error("list project invites", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list invites"})
		return
	}
	if invites == nil {
		invites = []*domain.ProjectInvite{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": invites})
}

func (h *Handler) createInvite(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	project, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer)
	if err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json body"})
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "email is required"})
		return
	}
	if domain.RoleRank(req.Role) < 0 {
		req.Role = domain.RoleDeveloper
	}
	// Only owners can invite with owner role.
	if req.Role == domain.RoleOwner {
		if _, err := h.authorizedProjectRole(r, projectID, domain.RoleOwner); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "only an owner can invite another owner"})
			return
		}
	}
	claims, _ := middleware.ClaimsFromContext(r.Context())

	// Validate that the email belongs to a registered user and get their ID.
	invitedMember, err := h.orgs.FindMemberByEmail(r.Context(), strings.TrimSpace(req.Email))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no registered user found with that email"})
		return
	}

	// Reject if they're already a member.
	if _, memberErr := h.repo.GetMember(r.Context(), projectID, invitedMember.UserID); memberErr == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user is already a member of this project"})
		return
	}

	// Reject if there's already a pending invite for this email.
	if _, activeErr := h.repo.FindActiveInviteByEmail(r.Context(), projectID, req.Email); activeErr == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "a pending invite already exists for this email"})
		return
	}

	inv := &domain.ProjectInvite{
		ProjectID:    projectID,
		InvitedEmail: req.Email,
		Role:         req.Role,
		InvitedBy:    claims.Subject,
	}
	if err := h.repo.CreateInvite(r.Context(), inv); err != nil {
		h.logger.Error("create project invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create invite"})
		return
	}

	if h.notifications != nil {
		_, _ = h.notifications.Create(r.Context(), application.CreateInput{
			UserID: invitedMember.UserID,
			Type:   notificationsdomain.TypeProjectInvitation,
			Title:  "You've been invited to a project",
			Body:   fmt.Sprintf("You've been invited to join project %s.", project.Name),
			Data:   map[string]any{"project_id": projectID, "invite_id": inv.ID, "token": inv.Token},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":            inv.ID,
		"project_id":    inv.ProjectID,
		"invited_email": inv.InvitedEmail,
		"role":          inv.Role,
		"token":         inv.Token,
		"expires_at":    inv.ExpiresAt,
		"created_at":    inv.CreatedAt,
	})
}

func (h *Handler) revokeInvite(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	inviteID := chi.URLParam(r, "inviteID")
	if _, err := h.authorizedProjectRole(r, projectID, domain.RoleMaintainer); err != nil {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
		return
	}
	if err := h.repo.RevokeInvite(r.Context(), projectID, inviteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite not found"})
			return
		}
		h.logger.Error("revoke invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to revoke invite"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resolveInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tokenHash := hashToken(token)
	inv, err := h.repo.FindInviteByToken(r.Context(), tokenHash)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite not found"})
		return
	}
	if inv.AcceptedAt != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "invite already accepted"})
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		writeJSON(w, http.StatusGone, map[string]string{"error": "invite expired"})
		return
	}
	project, _ := h.repo.FindByID(r.Context(), inv.ProjectID)
	resp := map[string]any{
		"id":            inv.ID,
		"project_id":    inv.ProjectID,
		"invited_email": inv.InvitedEmail,
		"role":          inv.Role,
		"expires_at":    inv.ExpiresAt,
	}
	if project != nil {
		resp["project_name"] = project.Name
		resp["project_slug"] = project.Slug
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tokenHash := hashToken(token)
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	inv, err := h.repo.AcceptInvite(r.Context(), tokenHash, claims.Subject, claims.Email, claims.Username)
	if err != nil {
		if strings.Contains(err.Error(), "already accepted") {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "invite already accepted"})
			return
		}
		if strings.Contains(err.Error(), "expired") {
			writeJSON(w, http.StatusGone, map[string]string{"error": "invite expired"})
			return
		}
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "invite not found"})
			return
		}
		h.logger.Error("accept project invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to accept invite"})
		return
	}

	if h.notifications != nil {
		// Mark the invite notification as read now that it has been accepted.
		_ = h.notifications.MarkReadByInviteID(r.Context(), claims.Subject, inv.ID)

		project, _ := h.repo.FindByID(r.Context(), inv.ProjectID)
		projectName := inv.ProjectID
		if project != nil {
			projectName = project.Name
		}
		_, _ = h.notifications.Create(r.Context(), application.CreateInput{
			UserID: claims.Subject,
			Type:   notificationsdomain.TypeProjectInvitation,
			Title:  "You joined a project",
			Body:   fmt.Sprintf("You have successfully joined %s.", projectName),
			Data:   map[string]any{"project_id": inv.ProjectID},
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"project_id": inv.ProjectID})
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveOrganizationID determines the caller's active organization.
//
// If X-Organization-ID / organization_id is present:
//   - verify the caller is a member of that org (membership table)
//
// Otherwise fall back to the caller's personal org (derived from JWT sub),
// bootstrapping the org row + owner membership on first use.
func (h *Handler) resolveOrganizationID(r *http.Request) (string, error) {
	queryOrgID := strings.TrimSpace(r.URL.Query().Get("organization_id"))
	headerOrgID := strings.TrimSpace(r.Header.Get("X-Organization-ID"))

	if queryOrgID != "" && headerOrgID != "" && queryOrgID != headerOrgID {
		return "", fmt.Errorf("forbidden: conflicting organization context")
	}
	requestedOrgID := queryOrgID
	if requestedOrgID == "" {
		requestedOrgID = headerOrgID
	}

	claims, hasClaims := middleware.ClaimsFromContext(r.Context())

	// Explicit org in request — verify membership.
	if requestedOrgID != "" {
		if hasClaims && claims.Subject != "" && h.orgs != nil {
			if _, err := h.orgs.GetMember(r.Context(), requestedOrgID, claims.Subject); err != nil {
				return "", fmt.Errorf("forbidden: not a member of this organization")
			}
		}
		return requestedOrgID, nil
	}

	// No explicit org — use personal org derived from JWT sub.
	if hasClaims && claims.Subject != "" {
		personalOrgID := "org-" + normalizeSlug(claims.Subject)
		if h.orgs != nil {
			orgName := "My Organization"
			if claims.Username != "" {
				orgName = claims.Username + "'s Organization"
			}
			if err := h.orgs.EnsureOrgWithOwner(r.Context(), personalOrgID, orgName,
				claims.Subject, claims.Email, claims.Username); err != nil {
				return "", fmt.Errorf("failed to bootstrap organization")
			}
		} else {
			// Fallback when org repo not available (tests/legacy).
			if err := h.repo.EnsureOrganization(r.Context(), personalOrgID, "Organization "+personalOrgID); err != nil {
				return "", fmt.Errorf("failed to ensure organization")
			}
		}
		return personalOrgID, nil
	}

	return "org-default", nil
}

func (h *Handler) authorizedProject(r *http.Request, projectID string) (*domain.Project, error) {
	project, err := h.repo.FindByID(r.Context(), projectID)
	if err != nil {
		return nil, err
	}

	orgID, err := h.resolveOrganizationID(r)
	if err != nil {
		return nil, err
	}

	// Org matches — access granted.
	if orgID == "" || project.OrganizationID == orgID {
		return project, nil
	}

	// Org doesn't match, but the caller may be an explicit project member
	// (e.g. an invited user whose personal org differs from the project owner's org).
	if claims, ok := middleware.ClaimsFromContext(r.Context()); ok && claims.Subject != "" {
		if _, memberErr := h.repo.GetMember(r.Context(), projectID, claims.Subject); memberErr == nil {
			return project, nil
		}
	}

	return nil, fmt.Errorf("forbidden: project does not belong to caller organization")
}

// authorizedProjectRole checks org membership AND project-level role.
// minRole is the minimum role required (viewer < developer < maintainer < owner).
func (h *Handler) authorizedProjectRole(r *http.Request, projectID, minRole string) (*domain.Project, error) {
	project, err := h.authorizedProject(r, projectID)
	if err != nil {
		return nil, err
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		return nil, fmt.Errorf("forbidden: unauthorized")
	}
	member, err := h.repo.GetMember(r.Context(), projectID, claims.Subject)
	if err != nil {
		// If no member row exists, fall back to treating org owners as project owners.
		if errors.Is(err, sql.ErrNoRows) {
			if h.orgs != nil {
				orgMember, orgErr := h.orgs.GetMember(r.Context(), project.OrganizationID, claims.Subject)
				if orgErr == nil && orgMember.Role == "owner" {
					return project, nil
				}
			}
			return nil, fmt.Errorf("forbidden: not a member of this project")
		}
		return nil, err
	}
	if domain.RoleRank(member.Role) < domain.RoleRank(minRole) {
		return nil, fmt.Errorf("forbidden: requires %s role or higher", minRole)
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
