package registration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Config holds the registration webhook URL.
//
// Set REGISTRATION_WEBHOOK_URL to the endpoint that creates users in your
// identity provider. The platform will forward the registration payload to
// that URL and relay its response back to the panel.
//
// The webhook receives:
//
//	POST <REGISTRATION_WEBHOOK_URL>
//	Content-Type: application/json
//	{ "username": "...", "email": "...", "password": "..." }
//
// It should respond with 2xx on success, or a non-2xx status with a JSON
// body containing an "error" field describing the failure.
//
// Examples of what the webhook can be:
//   - A small sidecar script you write for your IDP (Keycloak, Authentik, etc.)
//   - A serverless function
//   - The IDP's own admin API if it accepts this exact payload
type Config struct {
	WebhookURL string
}

// Handler exposes a single public endpoint for headless user registration.
// It is registered on the unauthenticated mux — no bearer token is required.
type Handler struct {
	cfg    Config
	logger *slog.Logger
}

func NewHandler(cfg Config, logger *slog.Logger) *Handler {
	return &Handler{cfg: cfg, logger: logger}
}

// RegisterRoutes registers the public registration endpoint.
// Call this on the root mux (NOT the authenticated apiMux).
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/register", h.handleRegister)
}

type registerRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	if h.cfg.WebhookURL == "" {
		writeJSON(w, http.StatusNotImplemented, map[string]string{
			"error": "headless registration is not configured; set REGISTRATION_WEBHOOK_URL",
		})
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Username == "" || req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "username, email, and password are required"})
		return
	}

	payload, _ := json.Marshal(req)

	webhookReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.cfg.WebhookURL, bytes.NewReader(payload))
	if err != nil {
		h.logger.Error("registration webhook: build request failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	webhookReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(webhookReq)
	if err != nil {
		h.logger.Error("registration webhook: request failed", "error", err, "url", h.cfg.WebhookURL)
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error": fmt.Sprintf("registration webhook unreachable: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	h.logger.Info("registration webhook response", "status", resp.StatusCode, "username", req.Username)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
