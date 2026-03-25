package http

import (
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/catalog/ports"
)

const basePath = "/api/v1/catalog"

type Handler struct {
	crates     ports.CrateRepository
	blueprints ports.BlueprintRepository
	logger     *slog.Logger
}

func NewHandler(crates ports.CrateRepository, blueprints ports.BlueprintRepository, logger *slog.Logger) *Handler {
	return &Handler{crates: crates, blueprints: blueprints, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/crates", h.listCrates)
	mux.HandleFunc("GET "+basePath+"/crates/{id}/blueprints", h.listBlueprints)
	mux.HandleFunc("GET "+basePath+"/blueprints/{id}", h.getBlueprint)
}

func (h *Handler) listCrates(w http.ResponseWriter, r *http.Request) {
	crates, err := h.crates.ListCrates(r.Context())
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.JSON(w, http.StatusOK, crates)
}

func (h *Handler) listBlueprints(w http.ResponseWriter, r *http.Request) {
	crateID := r.PathValue("id")
	bps, err := h.blueprints.ListBlueprints(r.Context(), crateID)
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.JSON(w, http.StatusOK, bps)
}

func (h *Handler) getBlueprint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	bp, err := h.blueprints.GetBlueprint(r.Context(), id)
	if err != nil {
		commonhttp.Error(w, domain.NewNotFound("blueprint "+id))
		return
	}
	commonhttp.JSON(w, http.StatusOK, bp)
}
