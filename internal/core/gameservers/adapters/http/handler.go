package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/gameservers/application/commands"
	gsports "github.com/kleff/platform/internal/core/gameservers/ports"
	"github.com/kleff/platform/internal/shared/middleware"
)

const basePath = "/api/v1/gameservers"

type Handler struct {
	provision *commands.ProvisionServerHandler
	stop      *commands.StopServerHandler
	repo      gsports.GameServerRepository
	logger    *slog.Logger
}

func NewHandler(
	provision *commands.ProvisionServerHandler,
	stop *commands.StopServerHandler,
	repo gsports.GameServerRepository,
	logger *slog.Logger,
) *Handler {
	return &Handler{provision: provision, stop: stop, repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath, h.list)
	mux.HandleFunc("POST "+basePath, h.create)
	mux.HandleFunc("GET "+basePath+"/{id}", h.get)
	mux.HandleFunc("POST "+basePath+"/{id}/stop", h.stop_)
}

// list returns all game servers for the caller's organization.
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("authentication required"))
		return
	}
	servers, err := h.repo.ListByOrg(r.Context(), claims.OrgID)
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}
	commonhttp.JSON(w, http.StatusOK, servers)
}

type createRequest struct {
	BlueprintID   string            `json:"blueprint_id"`
	Name          string            `json:"name"`
	EnvOverrides  map[string]string `json:"env_overrides,omitempty"`
	MemoryBytes   int64             `json:"memory_bytes,omitempty"`
	CPUMillicores int64             `json:"cpu_millicores,omitempty"`
}

// create provisions a new game server from a blueprint.
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("authentication required"))
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid request body"))
		return
	}
	if req.BlueprintID == "" {
		commonhttp.Error(w, domain.NewBadRequest("blueprint_id is required"))
		return
	}
	if req.Name == "" {
		commonhttp.Error(w, domain.NewBadRequest("name is required"))
		return
	}

	result, err := h.provision.Handle(r.Context(), commands.ProvisionServerCommand{
		OrganizationID: claims.OrgID,
		OwnerID:        claims.Subject,
		BlueprintID:    req.BlueprintID,
		Name:           req.Name,
		EnvOverrides:   req.EnvOverrides,
		MemoryBytes:    req.MemoryBytes,
		CPUMillicores:  req.CPUMillicores,
	})
	if err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}

	commonhttp.JSON(w, http.StatusCreated, map[string]string{"server_id": result.ServerID})
}

// get returns a single game server by ID.
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gs, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		commonhttp.Error(w, domain.NewNotFound("game server "+id))
		return
	}
	commonhttp.JSON(w, http.StatusOK, gs)
}

// stop_ enqueues a stop job for the daemon.
func (h *Handler) stop_(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("authentication required"))
		return
	}
	id := r.PathValue("id")

	if err := h.stop.Handle(r.Context(), commands.StopServerCommand{
		ServerID: id,
		OwnerID:  claims.Subject,
	}); err != nil {
		commonhttp.Error(w, domain.NewInternal(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
