package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

const basePath = "/api/v1/billing"

// Handler groups all HTTP endpoints for the billing module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all billing routes to the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get(basePath+"/subscription", h.getSubscription)
	r.Get(basePath+"/invoices", h.listInvoices)
	r.Get(basePath+"/invoices/{id}", h.getInvoice)
	r.Get(basePath+"/usage", h.getUsage)
	r.Post(basePath+"/portal", h.createPortalSession)
}

func notImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

func (h *Handler) getSubscription(w http.ResponseWriter, _ *http.Request)    { notImplemented(w) }
func (h *Handler) listInvoices(w http.ResponseWriter, _ *http.Request)       { notImplemented(w) }
func (h *Handler) getInvoice(w http.ResponseWriter, _ *http.Request)         { notImplemented(w) }
func (h *Handler) getUsage(w http.ResponseWriter, _ *http.Request)           { notImplemented(w) }
func (h *Handler) createPortalSession(w http.ResponseWriter, _ *http.Request) { notImplemented(w) }
