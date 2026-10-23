package routes

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type ImageRoutes struct {
	router
	imageService services.ImageService
	config       types.ImageConfig
}

func NewImageRoutes(
	imageService services.ImageService,
	config types.ImageConfig,
	router router) *ImageRoutes {
	return &ImageRoutes{
		router:       router,
		imageService: imageService,
		config:       config,
	}
}

// UploadImage parses the multipart request; the service owns upload validation and processing.
func (h *ImageRoutes) UploadImage(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	r.Body = http.MaxBytesReader(w, r.Body, int64(h.config.MaxFileSizeBytes)+(1<<20))
	if err := r.ParseMultipartForm(int64(h.config.MaxFileSizeBytes)); err != nil {
		u.RespondWithError(w, r, http.StatusRequestEntityTooLarge, err.Error())
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error retrieving file from form data")
		return
	}
	defer file.Close()

	var altText *string
	if value := r.FormValue("alt_text"); value != "" {
		altText = &value
	}

	images, err := h.imageService.Upload(r.Context(), services.ImageUpload{
		ProductID:        r.PathValue("id"),
		Filename:         header.Filename,
		File:             file,
		Type:             types.ParseImageType(r.FormValue("type")),
		AltText:          altText,
		RemoveBackground: r.URL.Query().Get("remove_bg") == "true",
		Enhance:          r.URL.Query().Get("enhance") == "true",
	})
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, "product not found")
		return
	}

	if errors.Is(err, services.ErrImageOptions) || errors.Is(err, services.ErrImageAIDisabled) {
		u.RespondWithJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	var providerErr *services.ImageProviderError
	if errors.As(err, &providerErr) {
		u.RespondWithJSON(w, http.StatusBadGateway, map[string]string{"message": "Image enhancement failed. The image was not uploaded. Please try again."})
		return
	}
	if errors.Is(err, services.ErrImageEnhancement) {
		slog.Warn("Image enhancement failed", "error", err)
		u.RespondWithJSON(w, http.StatusBadGateway, map[string]string{"message": "Image enhancement failed. The image was not uploaded. Please try again."})
		return
	}

	if errors.Is(err, services.ErrImageFileTooLarge) {
		u.RespondWithError(w, r, http.StatusRequestEntityTooLarge, services.ErrImageFileTooLarge.Error())
		return
	}

	if errors.Is(err, services.ErrImageResolution) {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, services.ErrImageResolution.Error())
		return
	}

	if errors.Is(err, services.ErrImageFormat) {
		u.RespondWithError(w, r, http.StatusUnsupportedMediaType, services.ErrImageFormat.Error())
		return
	}

	if errors.Is(err, services.ErrImageDecode) {
		u.RespondWithError(w, r, http.StatusInternalServerError, services.ErrImageDecode.Error())
		return
	}

	if errors.Is(err, services.ErrBackgroundRemoval) {
		u.RespondWithError(w, r, http.StatusBadGateway, "Background removal failed; please try uploading again")
		return
	}

	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, images)
}

func (h *ImageRoutes) RemoveImage(w http.ResponseWriter, r *http.Request) {
	err := h.imageService.Remove(r.Context(), r.PathValue("id"))
	if errors.Is(err, types.ErrNotFound) {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

func (h *ImageRoutes) RegisterRoutes() {
	h.mux.Handle("POST /images/products/{id}", h.secure(types.RoleAdmin)(h.UploadImage))
	h.mux.Handle("DELETE /images/{id}", h.secure(types.RoleAdmin)(h.RemoveImage))
}
