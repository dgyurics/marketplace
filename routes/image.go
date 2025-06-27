package routes

import (
	"net/http"
	"path/filepath"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/utilities"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type ImageRoutes struct {
	router
	imageService services.ImageService
}

func NewImageRoutes(imageService services.ImageService, router router) *ImageRoutes {
	return &ImageRoutes{
		router:       router,
		imageService: imageService,
	}
}

func (h *ImageRoutes) UploadImage(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"] // TODO verify product exists

	// Parse the multipart form data
	err := r.ParseMultipartForm(30 << 20) // 30 MB
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}
	file, fileHeader, err := r.FormFile("image") // "image" is the form field name for the file upload
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error retrieving file from form data")
		return
	}
	defer file.Close()

	// Validate file type
	supported, err := h.imageService.IsSupportedImage(file)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error checking image format")
		return
	}
	if !supported {
		u.RespondWithError(w, r, http.StatusUnsupportedMediaType, "unsupported image format")
		return
	}

	imgID, err := utilities.GenerateIDString()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image ID")
		return
	}

	// Extract file extension from original filename
	originalFilename := fileHeader.Filename
	ext := filepath.Ext(originalFilename) // Gets extension like ".jpg" or ".png"

	// Append extension to generated ID
	imageFilename := imgID + ext

	// Store the image
	imagePath, err := h.imageService.StoreImage(productID, file, imageFilename)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error storing image")
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"filename": imagePath,
	})
}

func (h *ImageRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/images/products/{id}", h.secureAdmin(h.UploadImage)).Methods(http.MethodPost)
}
