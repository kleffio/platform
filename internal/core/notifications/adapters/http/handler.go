// Package http exposes REST endpoints and an SSE stream for the notifications module.
package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kleffio/platform/internal/core/notifications/application"
	"github.com/kleffio/platform/internal/core/notifications/domain"
	"github.com/kleffio/platform/internal/shared/middleware"
)

const basePath = "/api/v1/notifications"

// Handler groups all HTTP endpoints for the notifications module.
type Handler struct {
	svc    *application.Service
	hub    *application.Hub
	logger *slog.Logger
}

// NewHandler creates a Handler.
func NewHandler(svc *application.Service, hub *application.Hub, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, hub: hub, logger: logger}
}

// RegisterRoutes attaches all notification routes to the provided router.
// All routes require a valid JWT (caller must already be wrapped in RequireAuth).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath, h.list)
	r.Get(basePath+"/unread-count", h.unreadCount)
	r.Get(basePath+"/stream", h.stream)
	r.Post(basePath+"/read-all", h.markAllRead)
	r.Patch(basePath+"/{id}/read", h.markRead)
	r.Delete(basePath+"/{id}", h.delete)
}

// ── List ──────────────────────────────────────────────────────────────────────

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	q := r.URL.Query()
	unreadOnly := q.Get("unread") == "true"
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	notifications, err := h.svc.List(r.Context(), claims.Subject, domain.ListFilter{
		UnreadOnly: unreadOnly,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		h.logger.Error("list notifications", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to list notifications"))
		return
	}
	if notifications == nil {
		notifications = []*domain.Notification{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}

// ── Unread count ──────────────────────────────────────────────────────────────

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	count, err := h.svc.CountUnread(r.Context(), claims.Subject)
	if err != nil {
		h.logger.Error("count unread notifications", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to count unread notifications"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

// ── Mark single read ──────────────────────────────────────────────────────────

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.MarkRead(r.Context(), id, claims.Subject); err != nil {
		h.logger.Error("mark notification read", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to mark notification as read"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Mark all read ─────────────────────────────────────────────────────────────

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	if err := h.svc.MarkAllRead(r.Context(), claims.Subject); err != nil {
		h.logger.Error("mark all notifications read", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to mark all notifications as read"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id, claims.Subject); err != nil {
		h.logger.Error("delete notification", "error", err)
		writeJSON(w, http.StatusInternalServerError, errBody("failed to delete notification"))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── SSE stream ────────────────────────────────────────────────────────────────

// stream opens a Server-Sent Events connection for the authenticated user.
// On connect it sends the current unread count, then streams new notifications
// as they arrive. A heartbeat comment is sent every 30 s to keep the connection
// alive through proxies.
func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	// Send initial connected event with the current unread count.
	count, _ := h.svc.CountUnread(r.Context(), claims.Subject)
	fmt.Fprintf(w, "event: connected\ndata: {\"unread_count\":%d}\n\n", count)
	flusher.Flush()

	ch := h.hub.Subscribe(claims.Subject)
	defer h.hub.Unsubscribe(claims.Subject, ch)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			// Heartbeat — keeps the connection alive through idle-timeout proxies.
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()

		case n, open := <-ch:
			if !open {
				return
			}
			data, err := json.Marshal(n)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
