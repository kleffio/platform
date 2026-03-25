// Package http wires together the profiles application layer and exposes it
// as a set of REST endpoints under /api/v1/users.
package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	commonhttp "github.com/kleff/go-common/adapters/http"
	"github.com/kleff/go-common/domain"
	"github.com/kleff/platform/internal/core/profiles/application/commands"
	"github.com/kleff/platform/internal/core/profiles/application/queries"
	profiledomain "github.com/kleff/platform/internal/core/profiles/domain"
	"github.com/kleff/platform/internal/shared/ids"
	"github.com/kleff/platform/internal/shared/middleware"
)

const (
	basePath      = "/api/v1/users"
	maxAvatarSize = 5 << 20 // 5 MiB
	uploadDir     = "uploads/avatars"
)

// Handler groups all HTTP endpoints for the profiles module.
type Handler struct {
	logger  *slog.Logger
	upsert  *commands.UpsertProfileHandler
	update  *commands.UpdateProfileHandler
	getProf *queries.GetProfileHandler
}

func NewHandler(
	logger *slog.Logger,
	upsert *commands.UpsertProfileHandler,
	update *commands.UpdateProfileHandler,
	getProf *queries.GetProfileHandler,
) *Handler {
	return &Handler{
		logger:  logger,
		upsert:  upsert,
		update:  update,
		getProf: getProf,
	}
}

// RegisterRoutes attaches all profile routes to the provided mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET "+basePath+"/me", h.getMe)
	mux.HandleFunc("PATCH "+basePath+"/me", h.updateMe)
	mux.HandleFunc("POST "+basePath+"/me/avatar", h.uploadAvatar)
}

// profileResponse is the JSON shape returned by GET and PATCH /me.
type profileResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username,omitempty"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	Bio             string `json:"bio,omitempty"`
	ThemePreference string `json:"theme_preference"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

func toResponse(p *profiledomain.UserProfile) profileResponse {
	return profileResponse{
		ID:              p.ID,
		Username:        p.Username,
		AvatarURL:       p.AvatarURL,
		Bio:             p.Bio,
		ThemePreference: string(p.ThemePreference),
		CreatedAt:       p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// GET /api/v1/users/me
//
// Lazy Creation entry point: if no profile exists yet for this Kratos identity,
// one is created with default values before the response is returned. This means
// there is no sign-up webhook needed — the first authenticated API call does the
// bootstrapping transparently.
func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}

	// UpsertProfile is idempotent: returns existing profile or creates a new one.
	// claims.Subject = Kratos identity.id = OIDC sub claim.
	profile, err := h.upsert.Handle(r.Context(), commands.UpsertProfileCommand{
		IdentityID: claims.Subject,
	})
	if err != nil {
		h.logger.Error("upsert profile", "identity_id", claims.Subject, "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("could not load profile")))
		return
	}

	commonhttp.Success(w, toResponse(profile))
}

// PATCH /api/v1/users/me
//
// Accepts a JSON body with any combination of: bio, theme_preference.
// Only provided fields are updated (pointer semantics in the command).
func (h *Handler) updateMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}

	var body struct {
		Bio             *string `json:"bio"`
		ThemePreference *string `json:"theme_preference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("invalid JSON body"))
		return
	}

	cmd := commands.UpdateProfileCommand{
		IdentityID: claims.Subject,
		Bio:        body.Bio,
	}
	if body.ThemePreference != nil {
		tp := profiledomain.ThemePreference(*body.ThemePreference)
		cmd.ThemePreference = &tp
	}

	profile, err := h.update.Handle(r.Context(), cmd)
	if err != nil {
		// Surface validation errors (e.g. invalid theme) as 400.
		if strings.Contains(err.Error(), "invalid theme_preference") {
			commonhttp.Error(w, domain.NewBadRequest(err.Error()))
			return
		}
		if strings.Contains(err.Error(), "profile not found") {
			commonhttp.Error(w, domain.NewNotFound("profile not found"))
			return
		}
		h.logger.Error("update profile", "identity_id", claims.Subject, "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("could not update profile")))
		return
	}

	commonhttp.Success(w, toResponse(profile))
}

// POST /api/v1/users/me/avatar
//
// Accepts multipart/form-data with a single "avatar" file field.
// Validates MIME type (image/*) and size (<= 5 MiB).
// Stores the file under uploads/avatars/<random-id>.<ext> and updates the profile.
//
// In production, replace the local file write with an S3 PutObject call and
// store the resulting presigned/public URL instead of a relative path.
func (h *Handler) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		commonhttp.Error(w, domain.NewUnauthorized("unauthorized"))
		return
	}

	// Limit the reader to prevent OOM on large uploads.
	r.Body = http.MaxBytesReader(w, r.Body, maxAvatarSize)
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		commonhttp.Error(w, domain.NewBadRequest("avatar must be under 5 MiB"))
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		commonhttp.Error(w, domain.NewBadRequest("missing avatar field in form"))
		return
	}
	defer file.Close()

	// Validate content type — only allow image/*.
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		commonhttp.Error(w, domain.NewBadRequest(fmt.Sprintf("unsupported content type %q: must be an image", contentType)))
		return
	}

	// Derive file extension from the MIME type or original filename.
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = mimeToExt(contentType)
	}

	// Ensure upload directory exists.
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		h.logger.Error("mkdir upload dir", "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("storage unavailable")))
		return
	}

	filename := ids.New() + ext
	destPath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(destPath)
	if err != nil {
		h.logger.Error("create avatar file", "path", destPath, "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("could not save avatar")))
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.logger.Error("write avatar file", "path", destPath, "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("could not save avatar")))
		return
	}

	// Public URL served by the API binary (or a CDN in front of it).
	// TODO: replace with an S3 URL when object storage is configured.
	avatarURL := "/static/avatars/" + filename

	profile, err := h.update.Handle(r.Context(), commands.UpdateProfileCommand{
		IdentityID: claims.Subject,
		AvatarURL:  &avatarURL,
	})
	if err != nil {
		h.logger.Error("update avatar_url", "identity_id", claims.Subject, "err", err)
		commonhttp.Error(w, domain.NewInternal(errors.New("could not update profile")))
		return
	}

	commonhttp.Success(w, toResponse(profile))
}

// mimeToExt returns a file extension for common image MIME types.
func mimeToExt(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".bin"
	}
}
