package routes

import (
	"context"
	"net/http"
	"path/filepath"

	"log/slog"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
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

const (
	formKeyImage   = "image"    // Form key for image file
	formKeyType    = "type"     // Form key for image type (e.g., "hero", "gallery", etc.)
	formKeyAltText = "alt_text" // Form key for alt text
)

// UploadImage handles the image upload for a product
// It verifies the product exists, checks the image format, stores the image on disk,
// generates signed URLs, and creates image records in the database.
//
// The image can be of type "hero", "gallery", or "thumbnail".
// The image is stored in a subdirectory named after the product ID.
//
// The image file is expected to be sent as a multipart form file with the key [image].
// The image [type] can be specified in the form data, defaulting to "gallery" if not provided.
// The [alt_text] can also be provided in the form data.
func (h *ImageRoutes) UploadImage(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]                       // product ID from path parameter
	removeBg := r.URL.Query().Get("remove_bg") == "true" // optional remove background flag

	// Parse the multipart form file
	if err := r.ParseMultipartForm(30 << 20); err != nil { // 30 MB limit
		u.RespondWithError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	// Verify product exists
	// FIXME: validate input is numeric
	exists, err := h.imageService.ProductExists(r.Context(), productID)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		u.RespondWithError(w, r, http.StatusNotFound, "product not found")
		return
	}

	// Retrieve the file from the form data
	file, fileHeader, err := r.FormFile(formKeyImage)
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error retrieving file from form data")
		return
	}
	defer file.Close()

	// Parse the multipart form image type
	imageType, _ := types.ParseImageType(r.FormValue(formKeyType))
	if imageType == "" {
		imageType = "gallery"
	}

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

	// Generate a unique ID for the file/image
	imgID, err := utilities.GenerateIDString()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image ID")
		return
	}

	// Extract file extension from original filename
	originalFilename := fileHeader.Filename
	ext := filepath.Ext(originalFilename) // Gets extension like ".jpg" or ".png"

	// Append extension to generated ID
	filename := imgID + ext

	// Store the image on disk
	imagePath, err := h.imageService.StoreImage(productID, file, filename)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error storing image")
		return
	}
	slog.Debug("Image uploaded successfully", "path", imagePath)

	// Generate signed URL(s) for the image
	// Note: If the image type is "hero", we generate URLs for hero, gallery, and thumbnail
	var urls []string
	if imageType == types.Hero {
		urls = h.imageService.CreateImageURLs(productID, filename, types.Hero, types.Gallery, types.Thumbnail)
	} else {
		urls = h.imageService.CreateImageURLs(productID, filename, imageType)
	}
	slog.Debug("Generated signed URL", "url", urls)

	// Create the image record(s)
	typs := []types.ImageType{imageType, types.Gallery, types.Thumbnail}
	for idx, url := range urls {
		id, err := utilities.GenerateIDString()
		if err != nil {
			u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image record ID")
			return
		}
		if err := h.imageService.CreateImageRecord(r.Context(), &types.Image{
			ID:        id,
			ProductID: productID,
			URL:       url,
			Type:      typs[idx],
			AltText:   altTextFromForm(r),
			Source:    filename,
		}); err != nil {
			slog.ErrorContext(r.Context(), "error creating image record", "productID", productID, "type", imageType, "error", err)
			u.RespondWithError(w, r, http.StatusInternalServerError, "error creating image record")
			return
		}
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{
		"path": imagePath,
	})

	// FIXME superfluous response.WriteHeader call
	// If background removal requested, do it asynchronously after response is sent
	if removeBg {
		// Clone the context to prevent it from being canceled when the request completes
		bgCtx := context.Background()
		go func() {
			newImagePath, err := h.imageService.RemoveBackground(bgCtx, imagePath, filename)
			if err != nil {
				slog.ErrorContext(r.Context(), "error removing background", "productID", productID, "error", err)
				u.RespondWithError(w, r, http.StatusInternalServerError, "error removing background")
				return
			}
			slog.Debug("Background removed successfully", "newPath", newImagePath)

			// Do we need to create new signatures/records??
		}()
	}
}

func altTextFromForm(r *http.Request) *string {
	altText := r.FormValue(formKeyAltText)
	if altText == "" {
		return nil
	}
	return &altText
}

func (h *ImageRoutes) RemoveImage(w http.ResponseWriter, r *http.Request) {
	err := h.imageService.RemoveImage(r.Context(), mux.Vars(r)["image"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "Image not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

// PromoteImage sets an images updated_at timestamp to the current time
// This is used to promote an image to the top of the gallery or thumbnail list
func (h *ImageRoutes) PromoteImage(w http.ResponseWriter, r *http.Request) {
	err := h.imageService.PromoteImage(r.Context(), mux.Vars(r)["image"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "Image not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

func (h *ImageRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/images/products/{id}", h.secureAdmin(h.UploadImage)).Methods(http.MethodPost)
	h.muxRouter.Handle("/images/{image}", h.secureAdmin(h.RemoveImage)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/images/{image}", h.secureAdmin(h.PromoteImage)).Methods(http.MethodPost)
}
