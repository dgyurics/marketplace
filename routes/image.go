package routes

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"log/slog"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type ImageRoutes struct {
	router
	imageTransformService services.ImageTransformService
	imageService          services.ImageService
	productService        services.ProductService
	imageConfig           types.ImageConfig
}

func NewImageRoutes(
	imageService services.ImageService,
	imageTransformService services.ImageTransformService,
	productService services.ProductService,
	config types.ImageConfig,
	router router) *ImageRoutes {
	return &ImageRoutes{
		router:                router,
		imageService:          imageService,
		imageTransformService: imageTransformService,
		productService:        productService,
		imageConfig:           config,
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
// 3. Optionally transforms the image (if transform=true and transformations are enabled)
// 4. Generates signed URLs and creates database records
//
// Request:
//
//	Form: multipart/form-data with "image" file field
//	Query: transform=true (optional, blocks until transformation completes)
//	Fields: type (hero|gallery|thumbnail, default: gallery), alt_text (optional)
//
// Response: 201 Created with image path
func (h *ImageRoutes) UploadImage(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	transform := r.URL.Query().Get("transform") == "true"

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

	// FIXME times out after 30 seconds. implement background running job + endpoint to poll status
	// Transform if requested; the service checks whether transformations are enabled.
	if transform {
		if _, err := h.imageTransformService.Transform(r.Context(), imagePath, filename); err != nil {
			slog.Error("error transforming image", "productID", productID, "imgPath", imagePath, "error", err)
			if err := os.Remove(imagePath); err != nil {
				slog.Error("error removing failed upload", "imgPath", imagePath, "error", err)
			}
			u.RespondWithError(w, r, http.StatusInternalServerError, "error transforming image")
			return
		}
	}

	// Create image records
	images, err := h.createImageRecords(w, r, productID, filename)
	if err != nil {
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, images)
}

func (h *ImageRoutes) parseAndValidateRequest(w http.ResponseWriter, r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	if err := r.ParseMultipartForm(int64(h.imageConfig.MaxFileSizeBytes)); err != nil {
		u.RespondWithError(w, r, http.StatusRequestEntityTooLarge, err.Error())
		return nil, nil, err
	}

	productID := r.PathValue("id")
	_, err := h.productService.GetProductByID(r.Context(), productID)
	if errors.Is(err, types.ErrNotFound) {
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
	err := h.imageService.ValidateImage(file)
	switch {
	case err == nil:
		return nil
	case errors.Is(err, types.ErrImageResolutionTooHigh):
		u.RespondWithError(w, r, http.StatusUnprocessableEntity, "image resolution too high")
	case errors.Is(err, types.ErrUnsupportedImageFormat):
		u.RespondWithError(w, r, http.StatusUnsupportedMediaType, "unsupported image format")
	default:
		u.RespondWithError(w, r, http.StatusInternalServerError, "error getting image resolution")
	}
	return err
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

func (h *ImageRoutes) createImageRecords(w http.ResponseWriter, r *http.Request, productID, filename string) ([]types.Image, error) {
	imageType := types.ParseImageType(r.FormValue(formKeyType))
	imageTypes := h.getImageTypes(imageType)
	urls := h.imageService.CreateImageURLs(productID, filename, imageTypes...)

	slog.Debug("Generated signed URLs", "count", len(urls))

	images := make([]types.Image, 0, len(urls))
	for idx, url := range urls {
		id, err := u.GenerateIDString()
		if err != nil {
			u.RespondWithError(w, r, http.StatusInternalServerError, "error generating image record ID")
			return nil, err
		}

		img := types.Image{
			ID:        id,
			ProductID: productID,
			URL:       url,
			Type:      imageTypes[idx],
			AltText:   altTextFromForm(r),
			Source:    filename,
		}
		if err := h.imageService.CreateImageRecord(r.Context(), &img); err != nil {
			slog.Error("error creating image record", "productID", productID, "type", imageTypes[idx], "error", err)
			u.RespondWithError(w, r, http.StatusInternalServerError, "error creating image record")
			return nil, err
		}
		images = append(images, img)
	}

	return images, nil
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
	err := h.imageService.RemoveImage(r.Context(), r.PathValue("image"))
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
	h.mux.Handle("DELETE /images/{image}", h.secure(types.RoleAdmin)(h.RemoveImage))
}
