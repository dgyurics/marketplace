package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

// ImageService handles image storage and produces signed transformation URLs
// that imgproxy trusts.
type ImageService interface {
	ValidateImage(file io.Reader) error
	StoreImage(productID string, file io.Reader, filename string) (string, error)
	CreateImageURLs(productID, filename string, imgType ...types.ImageType) []string
	CreateImageRecord(ctx context.Context, image *types.Image) error
	RemoveImage(ctx context.Context, id string) error
}

type imageService struct {
	repo           repositories.ImageRepository
	key            []byte // HMAC-SHA256 key shared with imgproxy for URL signing
	salt           []byte // Salt prepended to the message before signing
	baseURLImgPrxy string
	maxMegapixels  int
	imgDir         string
}

func NewImageService(repo repositories.ImageRepository, config types.ImageConfig) ImageService {
	return &imageService{
		repo:           repo,
		key:            config.Key,
		salt:           config.Salt,
		baseURLImgPrxy: config.BaseURLImgproxy,
		imgDir:         config.ImageUploadPath,
		maxMegapixels:  config.MaxMegapixels,
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
	// Copy the file content to the destination and close it even if copying fails.
	_, copyErr := io.Copy(dst, file)
	closeErr := dst.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}

	return filePath, nil
}

func (s *imageService) CreateImageRecord(ctx context.Context, image *types.Image) error {
	return s.repo.CreateImage(ctx, image)
}

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

// GenerateImageURL produces a signed imgproxy URL for the given image variant.
//
// HMAC-SHA256 is used because both the signer (this service) and the verifier
// (imgproxy) are internal infrastructure sharing a secret key. HMAC is simple,
// fast, and appropriate when there is no need to separate signing authority
// from verification.
//
// The signature prevents clients from crafting arbitrary transformation URLs,
// which could otherwise be used to abuse imgproxy.
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

	// signature = HMAC-SHA256(key, salt + path)
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

// ValidateImage checks the image format and resolution.
func (s *imageService) ValidateImage(file io.Reader) error {
	res, format, err := getImageInfo(file)
	if err != nil {
		return err
	}

	slog.Debug("Image resolution", "pixels", res)
	if res > s.maxMegapixels*1_000_000 {
		return types.ErrImageResolutionTooHigh
	}

	if !isSupportedFormat(format) {
		return fmt.Errorf("%w: %s", types.ErrUnsupportedImageFormat, format)
	}

	return nil
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
