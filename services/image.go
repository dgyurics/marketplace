package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

var (
	ErrImageFileTooLarge = errors.New("image file too large")
	ErrImageResolution   = errors.New("image resolution too high")
	ErrImageFormat       = errors.New("unsupported image format")
	ErrImageDecode       = errors.New("error getting image resolution")
	ErrBackgroundRemoval = errors.New("background removal failed")
	ErrImageOptions      = errors.New("select either remove background or enhance, not both")
	ErrImageAIDisabled   = errors.New("AI image enhancement is disabled")
	ErrImageEnhancement  = errors.New("image enhancement failed")
)

type ImageService interface {
	Upload(ctx context.Context, input ImageUpload) ([]types.Image, error)
	Remove(ctx context.Context, id string) error
}

type ImageUpload struct {
	ProductID        string
	Filename         string
	File             io.Reader
	Type             types.ImageType
	AltText          *string
	RemoveBackground bool
	Enhance          bool
}

type imageService struct {
	HttpClient    utilities.HTTPClient
	ai            types.AIModelConfig
	aiClient      *http.Client
	repo          repositories.ImageRepository
	products      ProductService
	config        types.ImageConfig
	baseURLRemBg  string
	maxFileSize   int64
	maxMegapixels int
}

func NewImageService(
	client utilities.HTTPClient,
	repo repositories.ImageRepository,
	products ProductService,
	config types.ImageConfig,
	ai types.AIModelConfig) ImageService {
	return &imageService{
		HttpClient: client,
		ai:         ai,
		aiClient: &http.Client{
			Timeout:       150 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
		},
		repo:          repo,
		products:      products,
		config:        config,
		baseURLRemBg:  config.BaseURLRembg,
		maxFileSize:   int64(config.MaxFileSizeBytes),
		maxMegapixels: config.MaxMegapixels,
	}
}

// Upload owns validation and the complete file-to-record lifecycle. The caller owns input.File.
func (s *imageService) Upload(ctx context.Context, input ImageUpload) ([]types.Image, error) {
	if input.RemoveBackground && input.Enhance {
		return nil, ErrImageOptions
	}
	if input.Enhance && !s.ai.Enabled {
		return nil, ErrImageAIDisabled
	}

	if _, err := s.products.GetProductByID(ctx, input.ProductID); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(io.LimitReader(input.File, s.maxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("error reading image: %w", err)
	}

	if int64(len(data)) > s.maxFileSize {
		return nil, ErrImageFileTooLarge
	}

	info, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrImageDecode, err)
	}

	if int64(info.Width)*int64(info.Height) > int64(s.maxMegapixels)*1_000_000 {
		return nil, ErrImageResolution
	}

	// currently, only jpeg & png supported
	if format != "jpeg" && format != "png" {
		return nil, ErrImageFormat
	}

	if input.Enhance {
		data, err = s.enhanceImage(ctx, data)
		input.Filename = "enhanced.png"
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrImageEnhancement, err)
	}

	imgID, err := utilities.GenerateIDString()
	if err != nil {
		return nil, fmt.Errorf("error generating image ID: %w", err)
	}

	filename := imgID + filepath.Ext(input.Filename)
	path, err := s.store(input.ProductID, bytes.NewReader(data), filename)
	if err != nil {
		return nil, fmt.Errorf("error storing image: %w", err)
	}

	if input.RemoveBackground {
		_, err = s.removeBackground(ctx, path, filename)
	}
	if err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("%w: %v", ErrBackgroundRemoval, err)
	}

	kind := types.ParseImageType(string(input.Type))
	kinds := []types.ImageType{kind}
	if kind == types.Hero {
		kinds = []types.ImageType{types.Hero, types.Gallery, types.Thumbnail}
	}

	images := make([]types.Image, 0, len(kinds))
	for _, kind := range kinds {
		id, err := utilities.GenerateIDString()
		if err != nil {
			return nil, fmt.Errorf("error generating image record ID: %w", err)
		}

		img := types.Image{
			ID:        id,
			ProductID: input.ProductID,
			Type:      kind,
			URL:       generateImageURL(s.config, input.ProductID, filename, kind),
			AltText:   input.AltText,
			Source:    filename,
		}
		if err := s.repo.CreateImage(ctx, &img); err != nil {
			return nil, fmt.Errorf("error creating image record: %w", err)
		}

		images = append(images, img)
	}

	return images, nil
}

// removeBackground sends the image at filePath to the rembg service for
// background removal, then overwrites the original file with the result.
func (s *imageService) removeBackground(ctx context.Context, filePath, filename string) (string, error) {
	// open source image
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// create multipart writer
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create form file using image
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy image into form file part
	if _, err := io.Copy(part, file); err != nil {
		return "", fmt.Errorf("failed to copy image file to form: %w", err)
	}

	// Add model field to form
	if err := writer.WriteField("model", "isnet-general-use"); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	// Close the multipart writer to finalize the form data
	writer.Close()

	// Prepare the HTTP request to the rembg service
	url := fmt.Sprintf("%s/api/remove", s.baseURLRemBg)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	res, err := s.HttpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to rembg service: %w", err)
	}
	defer res.Body.Close()

	// Handle response
	if res.StatusCode != http.StatusOK {
		slog.Error("Rembg service returned non-OK status", "status", res.StatusCode, "url", s.baseURLRemBg)
		return "", fmt.Errorf("failed to remove background: %s", res.Status)
	}

	// Overwrite the original image with the processed image
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the file content to the destination
	if _, err := io.Copy(dst, res.Body); err != nil {
		return "", err
	}

	return filePath, nil
}

func (s *imageService) Remove(ctx context.Context, imageID string) error {
	// remove image from database
	deleteResult, err := s.repo.RemoveImage(ctx, imageID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to remove image from DB", "error", err)
		return err
	}

	// if we cannot delete the source, return early
	if !deleteResult.CanDeleteSource {
		return nil
	}

	// remove image from filesystem
	directory := filepath.Join(s.config.ImageUploadPath, deleteResult.ProductID)
	filePath := filepath.Join(directory, deleteResult.SourceImage)
	if err := os.Remove(filePath); err != nil {
		slog.WarnContext(ctx, "Failed to remove image file", "file", filePath, "error", err)
		return err
	}

	// remove directory if empty
	entries, err := os.ReadDir(directory)
	if err != nil {
		slog.WarnContext(ctx, "Failed to read directory", "directory", directory, "error", err)
		return err
	}
	if len(entries) == 0 {
		if err := os.Remove(directory); err != nil {
			slog.WarnContext(ctx, "Failed to remove empty directory", "directory", directory, "error", err)
		}
	}

	return nil
}

// mkdir creates a directory for the image file, returning the full path to the file
func (s *imageService) mkdir(productID string, imagename string) (string, error) {
	dirPath := filepath.Join(s.config.ImageUploadPath, productID)
	filePath := filepath.Join(dirPath, filepath.Base(imagename))
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	return filePath, nil
}

// store saves the image file to disk and returns the file path
func (s *imageService) store(productID string, file io.Reader, filename string) (string, error) {
	filePath, err := s.mkdir(productID, filename)
	if err != nil {
		return "", fmt.Errorf("failed to create directory for image: %w", err)
	}

	// Save the file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the file content to the destination
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filePath, nil
}

const (
	// Image transformation settings
	GalleryResolution   = "resize:fit:1200:800:0"
	ThumbnailResolution = "resize:fit:300:300:0"
	HeroResolution      = "resize:fit:1600:1200:0"
	DefaultResolution   = "resize:fit:800:600:0"

	GalleryQuality   = "quality:85"
	ThumbnailQuality = "quality:80"
	HeroQuality      = "quality:90"
	DefaultQuality   = "quality:85"
)

// url produces a signed imgproxy URL for the given image variant.
//
// HMAC-SHA256 is used because both the signer (this service) and the verifier
// (imgproxy) are internal infrastructure sharing a secret key. HMAC is simple,
// fast, and appropriate when there is no need to separate signing authority
// from verification.
//
// The signature prevents clients from crafting arbitrary transformation URLs,
// which could otherwise be used to abuse imgproxy.
func generateImageURL(config types.ImageConfig, productID, filename string, imgType types.ImageType) string {
	var path string
	switch imgType {
	case types.Gallery:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s", GalleryResolution, GalleryQuality, productID, filename)
	case types.Thumbnail:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s", ThumbnailResolution, ThumbnailQuality, productID, filename)
	case types.Hero:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s", HeroResolution, HeroQuality, productID, filename)
	default:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s", DefaultResolution, DefaultQuality, productID, filename)
	}

	// Explicit PNG output prevents image delivery from flattening transparency.
	if strings.EqualFold(filepath.Ext(filename), ".png") {
		path = "/format:png" + path
	}

	// signature = HMAC-SHA256(key, salt + path)
	mac := hmac.New(sha256.New, config.Key)
	mac.Write(config.Salt)
	mac.Write([]byte(path))
	signature := mac.Sum(nil)
	encodedSig := base64.RawURLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s/%s%s", config.BaseURLImgproxy, encodedSig, path)
}
