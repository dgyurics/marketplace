package routes

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
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

// UploadImage uploads and processes an image for a product.
//
// Process:
// 1. Validates product exists and image format/resolution
// 2. Stores image to disk
// 3. Optionally removes background (if remove_bg=true)
// 4. Generates signed URLs and creates database records
//
// Request:
//
//	Form: multipart/form-data with "image" file field
//	Query: remove_bg=true (optional, blocks until background removal completes)
//	Fields: type (hero|gallery|thumbnail, default: gallery), alt_text (optional)
//
// Response: 201 Created with image path
func (h *ImageRoutes) UploadImage(w http.ResponseWriter, r *http.Request) {
	productID := mux.Vars(r)["id"]
	removeBg := r.URL.Query().Get("remove_bg") == "true"

	// Parse and validate request
	file, fileHeader, err := h.parseAndValidateRequest(w, r)
	if err != nil {
		return
	}
	defer file.Close()

	// Validate image
	if err := h.validateImage(w, r, file); err != nil {
		return
	}

	// Reset file reader
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error resetting file reader")
		return
	}

	// Store image
	filename, imagePath, err := h.storeImage(w, r, productID, file, fileHeader)
	if err != nil {
		return
	}

	// Remove background if requested
	if removeBg {
		if err := h.removeBackground(w, r, productID, imagePath, filename); err != nil {
			return
		}
	}

	// Create image records
	if err := h.createImageRecords(w, r, productID, filename); err != nil {
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, map[string]string{"path": imagePath})
}

func (h *ImageRoutes) parseAndValidateRequest(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(int64(h.config.MaxFileSizeBytes)); err != nil {
		u.RespondWithError(w, r, http.StatusRequestEntityTooLarge, err.Error())
		return nil, nil, err
	}

	productID := mux.Vars(r)["id"]
	_, err := h.productService.GetProductByID(r.Context(), productID)
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, "product not found")
		return nil, nil, err
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return nil, nil, err
	}

	file, fileHeader, err := r.FormFile(formKeyImage)
	if err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error retrieving file from form data")
		return nil, nil, err
	}

	return file, fileHeader, nil
}

func (h *ImageRoutes) validateImage(w http.ResponseWriter, r *http.Request, file io.Reader) error {
	res, format, err := getImageInfo(file)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error getting image resolution")
		return err
	}

	slog.Debug("Image resolution", "pixels", res)
	if res > h.config.MaxMegapixels*1_000_000 {
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, "image resolution too high")
		return fmt.Errorf("resolution too high")
	}

	if !isSupportedFormat(format) {
		u.RespondWithError(w, r, http.StatusUnsupportedMediaType, "unsupported image format")
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

func (h *ImageRoutes) storeImage(w http.ResponseWriter, r *http.Request, productID string, file io.Reader, fileHeader *multipart.FileHeader) (string, string, error) {
	imgID, err := u.GenerateIDString()
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image ID")
		return "", "", err
	}

	filename := imgID + filepath.Ext(fileHeader.Filename)
	imagePath, err := h.imageService.StoreImage(productID, file, filename)
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, "error storing image")
		return "", "", err
	}

	slog.Debug("Image stored successfully", "path", imagePath)
	return filename, imagePath, nil
}

func (h *ImageRoutes) removeBackground(w http.ResponseWriter, r *http.Request, productID, imagePath, filename string) error {
	bgCtx := context.Background()
	newImagePath, err := h.imageService.RemoveBackground(bgCtx, imagePath, filename)
	if err != nil {
		slog.Error("error removing background", "productID", productID, "imgPath", imagePath, "error", err)
		u.RespondWithError(w, r, http.StatusInternalServerError, "error removing background")
		return err
	}

	slog.Debug("Background removed successfully", "newPath", newImagePath)
	return nil
}

func (h *ImageRoutes) createImageRecords(w http.ResponseWriter, r *http.Request, productID, filename string) error {
	imageType := types.ParseImageType(r.FormValue(formKeyType))
	imageTypes := h.getImageTypes(imageType)
	urls := h.imageService.CreateImageURLs(productID, filename, imageTypes...)

	slog.Debug("Generated signed URLs", "count", len(urls))

	typesToCreate := []types.ImageType{imageType, types.Gallery, types.Thumbnail}
	for idx, url := range urls {
		id, err := u.GenerateIDString()
		if err != nil {
			u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image record ID")
			return err
		}

		if err := h.imageService.CreateImageRecord(r.Context(), &types.Image{
			ID:        id,
			ProductID: productID,
			URL:       url,
			Type:      typesToCreate[idx],
			AltText:   altTextFromForm(r),
			Source:    filename,
		}); err != nil {
			slog.Error("error creating image record", "productID", productID, "type", typesToCreate[idx], "error", err)
			u.RespondWithError(w, r, http.StatusInternalServerError, "error creating image record")
			return err
		}
	}

	return nil
}

func (h *ImageRoutes) getImageTypes(imageType types.ImageType) []types.ImageType {
	if imageType == types.Hero {
		return []types.ImageType{types.Hero, types.Gallery, types.Thumbnail}
	}
	return []types.ImageType{imageType}
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
