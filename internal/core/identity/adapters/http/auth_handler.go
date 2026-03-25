package http

import (
	"encoding/json"
	"log/slog"
	"net/http"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/identity/application/commands"
)

// AuthHandler exposes public endpoints for headless authentication.
// Routes are registered on the unauthenticated mux — no bearer token is required.
type AuthHandler struct {
	login    *commands.LoginHandler
	register *commands.RegisterHandler
	logger   *slog.Logger
}

// NewAuthHandler creates an AuthHandler wired to the given command handlers.
func NewAuthHandler(
	login *commands.LoginHandler,
	register *commands.RegisterHandler,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{login: login, register: register, logger: logger}
}

// RegisterPublicRoutes attaches the auth endpoints to the given (public, unauthenticated) mux.
func (h *AuthHandler) RegisterPublicRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", h.handleLogin)
	mux.HandleFunc("POST /api/v1/auth/register", h.handleRegister)
}

// ── Login ─────────────────────────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid request body"))
		return
	}
	if req.Username == "" || req.Password == "" {
		commonhttp.Error(w, domain.NewBadRequest("username and password are required"))
		return
	}

	result, err := h.login.Handle(r.Context(), commands.LoginCommand{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.logger.Warn("headless login failed", "username", req.Username, "error", err)
		commonhttp.Error(w, err)
		return
	}

	commonhttp.Success(w, result.Token)
}

// ── Register ──────────────────────────────────────────────────────────────────

type registerRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid request body"))
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		commonhttp.Error(w, domain.NewBadRequest("username, email, and password are required"))
		return
	}

	result, err := h.register.Handle(r.Context(), commands.RegisterCommand{
		Email:     req.Email,
		Username:  req.Username,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		h.logger.Warn("headless registration failed", "username", req.Username, "error", err)
		commonhttp.Error(w, err)
		return
	}

	commonhttp.Created(w, map[string]string{"user_id": result.UserID})
}
