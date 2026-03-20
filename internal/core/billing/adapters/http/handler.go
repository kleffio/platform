package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
)

const basePath = "/api/v1/billing"

// Handler groups all HTTP endpoints for the billing module.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

// RegisterRoutes attaches all billing routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/subscription", h.getSubscription)
	mux.HandleFunc("GET "+basePath+"/invoices", h.listInvoices)
	mux.HandleFunc("GET "+basePath+"/invoices/{id}", h.getInvoice)
	mux.HandleFunc("GET "+basePath+"/usage", h.getUsage)
	mux.HandleFunc("POST "+basePath+"/portal", h.createPortalSession)
}

func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

func (h *Handler) getUsage(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}

// createPortalSession creates a Stripe billing portal session.
func (h *Handler) createPortalSession(w http.ResponseWriter, r *http.Request) {
	commonhttp.Error(w, domain.NewUnauthorized("not implemented"))
}
