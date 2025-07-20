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

const (
	// Image upload path relative to the application root
	ImageUploadPath = "images"
)

type ImageService interface {
	StoreImage(productID string, file io.Reader, filename string) (string, error)
	IsSupportedImage(file io.Reader) (bool, error)
	CreateImageURLs(productID, filename string, imgType ...types.ImageType) []string
	CreateImageRecord(ctx context.Context, image *types.Image) error
	ProductExists(ctx context.Context, productID string) (bool, error) // TODO move to product service
	RemoveBackground(ctx context.Context, filePath, filename string) (string, error)
}

type imageService struct {
	HttpClient     utilities.HTTPClient
	repo           repositories.ImageRepository
	key            []byte
	salt           []byte
	baseURLImgPrxy string
	baseURLRemBg   string
}

func NewImageService(HttpClient utilities.HTTPClient, repo repositories.ImageRepository, config types.ImageConfig) ImageService {
	return &imageService{
		HttpClient:     HttpClient,
		repo:           repo,
		key:            config.Key,
		salt:           config.Salt,
		baseURLImgPrxy: config.BaseURLImgPrxy,
		baseURLRemBg:   config.BaseURLRemBg,
	}
}

func (s *imageService) ProductExists(ctx context.Context, productID string) (bool, error) {
	// TODO verify productID is number
	exists, err := s.repo.ProductExists(ctx, productID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// mkdir creates a directory for the image file, returning the full path to the file
func mkdir(productID string, imagename string) (string, error) {
	dirPath := filepath.Join(ImageUploadPath, productID)
	filePath := filepath.Join(dirPath, filepath.Base(imagename))
	err := os.MkdirAll(dirPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}
	return filePath, nil
}

// StoreImage saves the image file to disk and returns the file path
func (s *imageService) StoreImage(productID string, file io.Reader, filename string) (string, error) {
	filePath, err := mkdir(productID, filename)
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

func (s *imageService) IsSupportedImage(file io.Reader) (bool, error) {
	// Read the first 512 bytes to check the file type
	buff := make([]byte, 512)
	if _, err := file.Read(buff); err != nil {
		return false, err
	}

	// Reset the read pointer
	if seeker, ok := file.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return false, err
		}
	}

	// Detect content type
	contentType := http.DetectContentType(buff) // This will return a valid MIME type based on the first 512 bytes
	return isSupportedContentType(contentType), nil
}

// isSupportedContentType checks if the content type is one of the supported image formats
func isSupportedContentType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/bmp", "image/tiff":
		return true
	default:
		return false
	}
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
