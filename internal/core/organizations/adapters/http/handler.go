package http

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/notifications/application"
	notificationsdomain "github.com/kleffio/platform/internal/core/notifications/domain"
	"github.com/kleffio/platform/internal/core/organizations/adapters/persistence"
	"github.com/kleffio/platform/internal/core/organizations/domain"
	"github.com/kleffio/platform/internal/core/organizations/ports"
	"github.com/kleffio/platform/internal/shared/ids"
	"github.com/kleffio/platform/internal/shared/middleware"
)

const basePath = "/api/v1/organizations"

// Handler groups all HTTP endpoints for the organizations module.
type Handler struct {
	repo          ports.OrganizationRepository
	notifications *application.Service
	logger        *slog.Logger
}

func NewHandler(repo ports.OrganizationRepository, notifications *application.Service, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, notifications: notifications, logger: logger}
}

// RegisterRoutes attaches all organizations routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Post(basePath, h.create)
	r.Get(basePath+"/{id}", h.get)
	r.Patch(basePath+"/{id}", h.update)
	r.Delete(basePath+"/{id}", h.delete)

	// Members sub-resource
	r.Get(basePath+"/{id}/members", h.listMembers)
	r.Post(basePath+"/{id}/members", h.addMember)
	r.Patch(basePath+"/{id}/members/{userId}", h.updateMemberRole)
	r.Delete(basePath+"/{id}/members/{userId}", h.removeMember)

	// Invites sub-resource
	r.Get(basePath+"/{id}/invites", h.listInvites)
	r.Post(basePath+"/{id}/invites", h.createInvite)
	r.Delete(basePath+"/{id}/invites/{inviteId}", h.revokeInvite)

	// Public invite resolution + accept
	r.Get("/api/v1/invites/{token}", h.resolveInvite)
	r.Post("/api/v1/invites/{token}/accept", h.acceptInvite)
}

// ── List orgs the caller belongs to ──────────────────────────────────────────

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	orgs, err := h.repo.ListByUserID(r.Context(), claims.Subject)
	if err != nil {
		h.logger.Error("list orgs", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to list organizations"))
		return
	}
	if orgs == nil {
		orgs = []*domain.Organization{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"organizations": orgs})
}

// ── Create org ────────────────────────────────────────────────────────────────

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json body"))
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, errBody("name is required"))
		return
	}

	now := time.Now().UTC()
	org := &domain.Organization{
		ID:        ids.New(),
		Name:      strings.TrimSpace(req.Name),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.repo.Save(r.Context(), org); err != nil {
		h.logger.Error("create organization", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to create organization"))
		return
	}

	// Caller is automatically the first owner.
	member := &domain.Member{
		OrgID:       org.ID,
		UserID:      claims.Subject,
		Email:       claims.Email,
		DisplayName: claims.Username,
		Role:        domain.RoleOwner,
		CreatedAt:   now,
	}
	if err := h.repo.AddMember(r.Context(), member); err != nil {
		h.logger.Error("add owner after create", "error", err)
		// Non-fatal: org was created, membership may be retried.
	}

	writeJSON(w, http.StatusCreated, org)
}

// ── Get org ───────────────────────────────────────────────────────────────────

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	org, err := h.authorizedOrg(r, id, "")
	if err != nil {
		writeOrgError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

// ── Update org name ───────────────────────────────────────────────────────────

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json body"))
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeJSON(w, http.StatusBadRequest, errBody("name is required"))
		return
	}

	org := &domain.Organization{
		ID:        id,
		Name:      strings.TrimSpace(req.Name),
		UpdatedAt: time.Now().UTC(),
	}
	if err := h.repo.Update(r.Context(), org); err != nil {
		h.logger.Error("update organization", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to update organization"))
		return
	}
	writeJSON(w, http.StatusOK, org)
}

// ── Delete org ────────────────────────────────────────────────────────────────

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.authorizedOrg(r, id, domain.RoleOwner); err != nil {
		writeOrgError(w, err)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		h.logger.Error("delete organization", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to delete organization"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Members ───────────────────────────────────────────────────────────────────

func (h *Handler) listMembers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.authorizedOrg(r, id, ""); err != nil {
		writeOrgError(w, err)
		return
	}

	members, err := h.repo.ListMembers(r.Context(), id)
	if err != nil {
		h.logger.Error("list members", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to list members"))
		return
	}
	if members == nil {
		members = []*domain.Member{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"members": members})
}

func (h *Handler) addMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	var req struct {
		UserID      string `json:"user_id"`
		Email       string `json:"email"`
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json body"))
		return
	}
	if req.UserID == "" {
		writeJSON(w, http.StatusBadRequest, errBody("user_id is required"))
		return
	}
	role := normalizeRole(req.Role)

	member := &domain.Member{
		OrgID:       id,
		UserID:      req.UserID,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Role:        role,
		CreatedAt:   time.Now().UTC(),
	}
	if err := h.repo.AddMember(r.Context(), member); err != nil {
		h.logger.Error("add member", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to add member"))
		return
	}
	writeJSON(w, http.StatusCreated, member)
}

func (h *Handler) updateMemberRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")

	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json body"))
		return
	}
	role := normalizeRole(req.Role)

	if err := h.repo.UpdateMemberRole(r.Context(), id, userID, role); err != nil {
		h.logger.Error("update member role", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to update role"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"role": role})
}

func (h *Handler) removeMember(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	// Members may remove themselves; removing others requires admin/owner.
	if claims.Subject != userID {
		if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
			writeOrgError(w, err)
			return
		}
	} else {
		// Self-removal: still need to be a member.
		if _, err := h.authorizedOrg(r, id, ""); err != nil {
			writeOrgError(w, err)
			return
		}
	}

	// Prevent removing the last owner.
	target, err := h.repo.GetMember(r.Context(), id, userID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody("member not found"))
		return
	}
	if target.Role == domain.RoleOwner {
		count, err := h.repo.CountOwners(r.Context(), id)
		if err != nil || count <= 1 {
			writeJSON(w, http.StatusConflict, errBody("cannot remove the last owner"))
			return
		}
	}

	if err := h.repo.RemoveMember(r.Context(), id, userID); err != nil {
		h.logger.Error("remove member", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to remove member"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Invites ───────────────────────────────────────────────────────────────────

func (h *Handler) listInvites(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	invites, err := h.repo.ListInvites(r.Context(), id)
	if err != nil {
		h.logger.Error("list invites", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to list invites"))
		return
	}
	if invites == nil {
		invites = []*domain.Invite{}
	}
	// Strip token hashes from the list response.
	type safeInvite struct {
		ID           string     `json:"id"`
		OrgID        string     `json:"org_id"`
		InvitedEmail string     `json:"invited_email"`
		Role         string     `json:"role"`
		InvitedBy    string     `json:"invited_by"`
		ExpiresAt    time.Time  `json:"expires_at"`
		AcceptedAt   *time.Time `json:"accepted_at,omitempty"`
		CreatedAt    time.Time  `json:"created_at"`
	}
	out := make([]safeInvite, len(invites))
	for i, inv := range invites {
		out[i] = safeInvite{
			ID:           inv.ID,
			OrgID:        inv.OrgID,
			InvitedEmail: inv.InvitedEmail,
			Role:         inv.Role,
			InvitedBy:    inv.InvitedBy,
			ExpiresAt:    inv.ExpiresAt,
			AcceptedAt:   inv.AcceptedAt,
			CreatedAt:    inv.CreatedAt,
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"invites": out})
}

func (h *Handler) createInvite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json body"))
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		writeJSON(w, http.StatusBadRequest, errBody("email is required"))
		return
	}

	token, err := generateToken()
	if err != nil {
		h.logger.Error("generate invite token", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to generate invite"))
		return
	}

	now := time.Now().UTC()
	inv := &domain.Invite{
		ID:           ids.New(),
		OrgID:        id,
		InvitedEmail: strings.ToLower(strings.TrimSpace(req.Email)),
		Role:         normalizeRole(req.Role),
		Token:        token,
		TokenHash:    persistence.HashToken(token),
		InvitedBy:    claims.Subject,
		ExpiresAt:    now.Add(7 * 24 * time.Hour),
		CreatedAt:    now,
	}

	if err := h.repo.CreateInvite(r.Context(), inv); err != nil {
		h.logger.Error("create invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to create invite"))
		return
	}

	// Notify the inviting admin that the invite was dispatched.
	if h.notifications != nil {
		_, _ = h.notifications.Create(r.Context(), application.CreateInput{
			UserID: claims.Subject,
			Type:   notificationsdomain.TypeOrgInvitation,
			Title:  "Invitation sent",
			Body:   fmt.Sprintf("An invitation was sent to %s to join the organization.", inv.InvitedEmail),
			Data:   map[string]any{"org_id": id, "invite_id": inv.ID, "invited_email": inv.InvitedEmail},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":            inv.ID,
		"org_id":        inv.OrgID,
		"invited_email": inv.InvitedEmail,
		"role":          inv.Role,
		"token":         inv.Token, // raw token returned once; client builds the accept URL
		"expires_at":    inv.ExpiresAt,
		"created_at":    inv.CreatedAt,
	})
}

func (h *Handler) revokeInvite(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	inviteID := chi.URLParam(r, "inviteId")

	if _, err := h.authorizedOrg(r, id, domain.RoleAdmin); err != nil {
		writeOrgError(w, err)
		return
	}

	if err := h.repo.RevokeInvite(r.Context(), inviteID); err != nil {
		h.logger.Error("revoke invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to revoke invite"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Public invite resolution ──────────────────────────────────────────────────

func (h *Handler) resolveInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tokenHash := persistence.HashToken(token)

	inv, err := h.repo.FindInviteByToken(r.Context(), tokenHash)
	if err == sql.ErrNoRows || err != nil {
		writeJSON(w, http.StatusNotFound, errBody("invite not found or expired"))
		return
	}
	if inv.AcceptedAt != nil {
		writeJSON(w, http.StatusGone, errBody("invite already accepted"))
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		writeJSON(w, http.StatusGone, errBody("invite has expired"))
		return
	}

	org, err := h.repo.FindByID(r.Context(), inv.OrgID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody("failed to load organization"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":            inv.ID,
		"org_id":        inv.OrgID,
		"org_name":      org.Name,
		"invited_email": inv.InvitedEmail,
		"role":          inv.Role,
		"invited_by":    inv.InvitedBy,
		"expires_at":    inv.ExpiresAt,
	})
}

func (h *Handler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	tokenHash := persistence.HashToken(token)

	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	inv, err := h.repo.FindInviteByToken(r.Context(), tokenHash)
	if err == sql.ErrNoRows || err != nil {
		writeJSON(w, http.StatusNotFound, errBody("invite not found or expired"))
		return
	}
	if inv.AcceptedAt != nil {
		writeJSON(w, http.StatusGone, errBody("invite already accepted"))
		return
	}
	if time.Now().After(inv.ExpiresAt) {
		writeJSON(w, http.StatusGone, errBody("invite has expired"))
		return
	}

	if err := h.repo.AcceptInvite(r.Context(), inv.ID, claims.Subject, claims.Email, claims.Username); err != nil {
		h.logger.Error("accept invite", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to accept invite"))
		return
	}

	// Notify the new member that they joined the organization.
	if h.notifications != nil {
		org, orgErr := h.repo.FindByID(r.Context(), inv.OrgID)
		orgName := inv.OrgID
		if orgErr == nil {
			orgName = org.Name
		}
		_, _ = h.notifications.Create(r.Context(), application.CreateInput{
			UserID: claims.Subject,
			Type:   notificationsdomain.TypeOrgInvitation,
			Title:  "You joined an organization",
			Body:   fmt.Sprintf("You have successfully joined %s.", orgName),
			Data:   map[string]any{"org_id": inv.OrgID},
		})
	}

	writeJSON(w, http.StatusOK, map[string]string{"org_id": inv.OrgID})
}

// ── Access guard ──────────────────────────────────────────────────────────────

// authorizedOrg loads the org and verifies the caller is a member.
// If minRole is non-empty, it also enforces the minimum role (owner > admin > member).
//
// When the caller has no membership row for their personal org (org-<slug>),
// we bootstrap it automatically so first-time access works without a separate
// project-list round-trip.
func (h *Handler) authorizedOrg(r *http.Request, orgID, minRole string) (*domain.Organization, error) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		return nil, fmt.Errorf("unauthorized")
	}

	org, err := h.repo.FindByID(r.Context(), orgID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("not found")
	}
	if err != nil {
		return nil, fmt.Errorf("internal")
	}

	member, err := h.repo.GetMember(r.Context(), orgID, claims.Subject)
	if err == sql.ErrNoRows {
		// Bootstrap the personal org membership row if this is the caller's own org.
		if personalOrgID(claims.Subject) == orgID {
			orgName := "My Organization"
			if claims.Username != "" {
				orgName = claims.Username + "'s Organization"
			}
			if bootstrapErr := h.repo.EnsureOrgWithOwner(r.Context(), orgID, orgName,
				claims.Subject, claims.Email, claims.Username); bootstrapErr != nil {
				h.logger.Error("bootstrap org membership", "error", bootstrapErr)
				return nil, fmt.Errorf("internal")
			}
			member, err = h.repo.GetMember(r.Context(), orgID, claims.Subject)
			if err != nil {
				return nil, fmt.Errorf("internal")
			}
		} else {
			return nil, fmt.Errorf("forbidden")
		}
	} else if err != nil {
		return nil, fmt.Errorf("internal")
	}

	if minRole != "" && !roleAtLeast(member.Role, minRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return org, nil
}

// personalOrgID returns the personal org ID for a given JWT subject,
// matching the derivation used in the projects handler.
func personalOrgID(subject string) string {
	s := strings.ToLower(strings.TrimSpace(subject))
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")
	// Keep only lowercase alphanum and hyphens.
	var b strings.Builder
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			b.WriteRune(c)
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	return "org-" + slug
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func roleAtLeast(have, need string) bool {
	rank := map[string]int{
		domain.RoleMember: 1,
		domain.RoleAdmin:  2,
		domain.RoleOwner:  3,
	}
	return rank[have] >= rank[need]
}

func normalizeRole(r string) string {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case domain.RoleOwner:
		return domain.RoleOwner
	case domain.RoleAdmin:
		return domain.RoleAdmin
	default:
		return domain.RoleMember
	}
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func writeOrgError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch msg {
	case "unauthorized":
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
	case "forbidden":
		writeJSON(w, http.StatusForbidden, errBody("forbidden"))
	case "not found":
		writeJSON(w, http.StatusNotFound, errBody("organization not found"))
	default:
		writeJSON(w, http.StatusInternalServerError, errBody("internal error"))
	}
}

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
