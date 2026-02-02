package routes

import (
	"context"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"

	"log/slog"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type ImageRoutes struct {
	router
	imageService   services.ImageService
	productService services.ProductService
	config         types.ImageConfig
}

func NewImageRoutes(
	imageService services.ImageService,
	productService services.ProductService,
	config types.ImageConfig,
	router router) *ImageRoutes {
	return &ImageRoutes{
		router:         router,
		imageService:   imageService,
		productService: productService,
		config:         config,
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
	if err := r.ParseMultipartForm(int64(h.config.MaxFileSizeBytes)); err != nil {
		u.RespondWithError(w, r, http.StatusRequestEntityTooLarge, err.Error())
		return
	}

	// Verify product exists
	_, err := h.productService.GetProductByID(r.Context(), productID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	// Retrieve the file from the form data
	file, fileHeader, err := r.FormFile(formKeyImage)
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error retrieving file from form data")
		return
	}
	defer file.Close()

	// Ensure image resolution does not exceed img proxy limit (200 megapixels)
	res, format, err := getImageInfo(file)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error getting image resolution")
		return
	}
	slog.Debug("Image resolution", "pixels", res)
	if res > h.config.MaxMegapixels*1_000_000 {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, "image resolution too high")
		return
	}
	if !isSupportedFormat(format) {
		u.RespondWithError(w, r, http.StatusUnsupportedMediaType, "unsupported image format")
		return
	}

	// Reset file reader to the beginning after reading
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error resetting file reader")
		return
	}

	// Generate a unique ID for the file/image
	imgID, err := u.GenerateIDString()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image ID")
		return
	}

	// Construct the filename with the original file extension
	originalFilename := fileHeader.Filename
	ext := filepath.Ext(originalFilename) // Gets extension like ".jpg" or ".png"
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
	imageType := types.ParseImageType(r.FormValue(formKeyType))
	if imageType == types.Hero {
		urls = h.imageService.CreateImageURLs(productID, filename, types.Hero, types.Gallery, types.Thumbnail)
	} else {
		urls = h.imageService.CreateImageURLs(productID, filename, imageType)
	}
	slog.Debug("Generated signed URL", "url", urls)

	// Create the image record(s)
	typs := []types.ImageType{imageType, types.Gallery, types.Thumbnail}
	for idx, url := range urls {
		id, err := u.GenerateIDString()
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

	// If background removal requested, do it asynchronously after response is sent
	if removeBg {
		go func() {
			// Clone the context to prevent it from being canceled when the request completes
			bgCtx := context.Background()
			newImagePath, err := h.imageService.RemoveBackground(bgCtx, imagePath, filename)
			if err != nil {
				slog.ErrorContext(bgCtx, "error removing background", "productID", productID, "imgPath", imagePath, "error", err)
				return
			}
			slog.Debug("Background removed successfully", "newPath", newImagePath)
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
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
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
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondSuccess(w)
}

// getImageInfo returns the total pixel count and format of the image from the reader
// e.g., for a jpeg image of 1920x1080, returns (2073600, "jpeg", nil)
// e.g., for a png image of 800x600, returns (480000, "png", nil)
func getImageInfo(r io.Reader) (int, string, error) {
	config, format, err := image.DecodeConfig(r)
	if err != nil {
		return 0, "", err
	}
	return config.Width * config.Height, format, nil
}

func isSupportedFormat(format string) bool {
	switch format {
	// "heic" <- HEIC support disabled for now
	// "webp" <- WEBP support disabled for now
	// "gif" <- GIF support disabled for now
	case "jpeg", "png":
		return true
	default:
		return false
	}
}

func (h *ImageRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/images/products/{id}", h.secure(types.RoleAdmin)(h.UploadImage)).Methods(http.MethodPost)
	h.muxRouter.Handle("/images/{image}", h.secure(types.RoleAdmin)(h.RemoveImage)).Methods(http.MethodDelete)
	h.muxRouter.Handle("/images/{image}", h.secure(types.RoleAdmin)(h.PromoteImage)).Methods(http.MethodPost)
}
