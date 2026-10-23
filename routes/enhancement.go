package routes

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

// EnhancementRoutes generates previews synchronously and accepts or discards them.
type EnhancementRoutes struct {
	router
	improvementService services.EnhancementService
}

func NewEnhancementRoutes(
	improvementService services.EnhancementService,
	router router) *EnhancementRoutes {
	return &EnhancementRoutes{
		router:             router,
		improvementService: improvementService,
	}
}

// GenerateEnhancement waits for a completed preview.
// The saved image stays unchanged until the preview is accepted.
func (h *EnhancementRoutes) GenerateEnhancement(w http.ResponseWriter, r *http.Request) {
	// Image generation can exceed the server's normal response deadline.
	if err := http.NewResponseController(w).SetWriteDeadline(time.Now().Add(180 * time.Second)); err != nil && !errors.Is(err, http.ErrNotSupported) {
		slog.Warn("Could not extend image enhancement response deadline", "error", err)
	}

	var body struct {
		Context string `json:"context"`
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&body)
	if err != nil && !errors.Is(err, io.EOF) {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid image context")
		return
	}

	var extra any
	if decoder.Decode(&extra) != io.EOF {
		u.RespondWithError(w, r, http.StatusBadRequest, "Invalid image context")
		return
	}

	preview, err := h.improvementService.Generate(r.Context(), r.PathValue("id"), body.Context)
	if errors.Is(err, services.ErrImageContextInvalid) {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, "Image or preview no longer exists")
		return
	}

	if errors.Is(err, services.ErrImageAIDisabled) {
		u.RespondWithError(w, r, http.StatusNotFound, "Image improvement is unavailable")
		return
	}

	var providerErr *services.ImageProviderError
	if errors.As(err, &providerErr) {
		u.RespondWithJSON(w, http.StatusBadGateway, map[string]string{
			"message": providerErr.Error() + " Your current image is unchanged.",
		})
		return
	}

	if errors.Is(err, services.ErrInvalidOriginalImage) {
		u.RespondWithError(w, r, http.StatusInternalServerError, "The original image could not be read. Upload a valid JPEG or PNG and try again.")
		return
	}

	if err != nil {
		slog.Warn("Image enhancement failed", "image_id", r.PathValue("id"), "error", err)
		u.RespondWithError(w, r, http.StatusInternalServerError, "Image processing failed; your image is unchanged")
		return
	}

	u.RespondWithJSON(w, http.StatusOK, preview)
}

func (h *EnhancementRoutes) ResolveEnhancement(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Accept *bool `json:"accept"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024)).Decode(&body); err != nil || body.Accept == nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "accept must be true or false")
		return
	}

	var err error
	if *body.Accept {
		err = h.improvementService.Accept(r.Context(), r.PathValue("id"), r.PathValue("preview_id"))
	} else {
		err = h.improvementService.Discard(r.Context(), r.PathValue("id"), r.PathValue("preview_id"))
	}

	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, "Image or preview no longer exists")
		return
	}

	if errors.Is(err, services.ErrImagePreviewStale) {
		u.RespondWithError(w, r, http.StatusConflict, services.ErrImagePreviewStale.Error())
		return
	}

	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "Image processing failed; your image is unchanged")
		return
	}

	u.RespondSuccess(w)
}

func (h *EnhancementRoutes) RegisterRoutes() {
	h.mux.Handle("POST /enhancements/images/{id}", h.secure(types.RoleAdmin)(h.GenerateEnhancement))
	h.mux.Handle("POST /enhancements/images/{id}/{preview_id}", h.secure(types.RoleAdmin)(h.ResolveEnhancement))
}
