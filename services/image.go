package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ImageService interface {
	StoreImage(productID string, file io.Reader, filename string) (string, error)
	CreateImageURLs(productID, filename string, imgType ...types.ImageType) []string
	CreateImageRecord(ctx context.Context, image *types.Image) error
	RemoveBackground(ctx context.Context, filePath, filename string) (string, error)
	RemoveImage(ctx context.Context, id string) error
}

type imageService struct {
	HttpClient     utilities.HTTPClient
	repo           repositories.ImageRepository
	key            []byte
	salt           []byte
	baseURLImgPrxy string
	baseURLRemBg   string
	imgDir         string
}

func NewImageService(HttpClient utilities.HTTPClient, repo repositories.ImageRepository, config types.ImageConfig) ImageService {
	return &imageService{
		HttpClient:     HttpClient,
		repo:           repo,
		key:            config.Key,
		salt:           config.Salt,
		baseURLImgPrxy: config.BaseURLImgproxy,
		baseURLRemBg:   config.BaseURLRembg,
		imgDir:         config.ImageUploadPath,
	}
}

// mkdir creates a directory for the image file, returning the full path to the file
func (s *imageService) mkdir(productID string, imagename string) (string, error) {
	dirPath := filepath.Join(s.imgDir, productID)
	filePath := filepath.Join(dirPath, filepath.Base(imagename))
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	return filePath, nil
}

// StoreImage saves the image file to disk and returns the file path
func (s *imageService) StoreImage(productID string, file io.Reader, filename string) (string, error) {
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

func (s *imageService) CreateImageRecord(ctx context.Context, image *types.Image) error {
	return s.repo.CreateImage(ctx, image)
}

// CreateImageURLs generates signed URLs for the specified image types
func (s *imageService) CreateImageURLs(productID, filename string, imgType ...types.ImageType) []string {
	urls := make([]string, len(imgType))
	for i, t := range imgType {
		urls[i] = s.GenerateImageURL(productID, filename, t)
	}
	return urls
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

// RemoveBackground removes the background from the image specified by imagePath
// It does so by sending a http multipart request to rembg service
func (s *imageService) RemoveBackground(ctx context.Context, filePath, filename string) (string, error) {
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

// GenerateImageURL generates a signed URL for use with imgproxy
func (s *imageService) GenerateImageURL(productID, filename string, imgType types.ImageType) string {
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

	mac := hmac.New(sha256.New, s.key)
	mac.Write(s.salt)
	mac.Write([]byte(path))
	signature := mac.Sum(nil)
	encodedSig := base64.RawURLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s/%s%s", s.baseURLImgPrxy, encodedSig, path)
}

func (s *imageService) RemoveImage(ctx context.Context, imageID string) error {
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
	directory := filepath.Join(s.imgDir, deleteResult.ProductID)
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
