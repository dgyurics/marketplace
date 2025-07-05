package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
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
}

type imageService struct {
	repo    repositories.ImageRepository
	key     []byte
	salt    []byte
	baseURL string
}

func NewImageService(repo repositories.ImageRepository, config types.ImageConfig) ImageService {
	return &imageService{
		repo:    repo,
		key:     config.Key,
		salt:    config.Salt,
		baseURL: config.BaseURL,
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

func (s *imageService) StoreImage(productID string, file io.Reader, filename string) (string, error) {
	// Create subdirectory for the file
	dirPath := filepath.Join(ImageUploadPath, productID)
	filePath := filepath.Join(dirPath, filepath.Base(filename))
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", err
	}

	// Save the file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

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

// GenerateImageURL generates a signed URL for use with imgproxy
func (s *imageService) GenerateImageURL(productID, filename string, imgType types.ImageType) string {
	var path string
	switch imgType {
	case types.Gallery:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s@webp", GalleryResolution, GalleryQuality, productID, filename)
	case types.Thumbnail:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s@webp", ThumbnailResolution, ThumbnailQuality, productID, filename)
	case types.Hero:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s@webp", HeroResolution, HeroQuality, productID, filename)
	default:
		path = fmt.Sprintf("/%s/%s/plain/local:///%s/%s@webp", DefaultResolution, DefaultQuality, productID, filename)
	}

	mac := hmac.New(sha256.New, s.key)
	mac.Write(s.salt)
	mac.Write([]byte(path))
	signature := mac.Sum(nil)
	encodedSig := base64.RawURLEncoding.EncodeToString(signature)

	return fmt.Sprintf("%s/%s%s", s.baseURL, encodedSig, path)
}
