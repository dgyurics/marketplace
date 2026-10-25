package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/dgyurics/marketplace/types"
)

type ImageTransformService interface {
	Transform(ctx context.Context, filePath, filename string) (string, error)
}

type imageTransformService struct {
	client *http.Client
	config types.ImageTransformConfig
}

func NewImageTransformService(client *http.Client, config types.ImageTransformConfig) ImageTransformService {
	return &imageTransformService{client: client, config: config}
}

// Transform processes the image when image transformations are enabled.
func (s *imageTransformService) Transform(ctx context.Context, filePath, filename string) (string, error) {
	if !s.config.Enabled {
		return filePath, nil
	}
	return s.removeBackground(ctx, filePath, filename)
}

// removeBackground sends the image at filePath to the rembg service for
// background removal, then overwrites the original file with the result.
func (s *imageTransformService) removeBackground(ctx context.Context, filePath, filename string) (string, error) {
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
	if err := writer.WriteField("model", s.config.Model); err != nil {
		return "", fmt.Errorf("failed to write model field: %w", err)
	}

	// Close the multipart writer to finalize the form data
	writer.Close()

	// Prepare the HTTP request to the rembg service
	url := fmt.Sprintf("%s/api/remove", s.config.BaseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Execute request
	res, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to rembg service: %w", err)
	}
	defer res.Body.Close()

	// Handle response
	if res.StatusCode != http.StatusOK {
		slog.Error("Rembg service returned non-OK status", "status", res.StatusCode, "url", s.config.BaseURL)
		return "", fmt.Errorf("failed to remove background: %s", res.Status)
	}

	// Overwrite the original image with the processed image
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}

	// Copy the file content to the destination and close it even if copying fails.
	_, copyErr := io.Copy(dst, res.Body)
	closeErr := dst.Close()
	if copyErr != nil {
		return "", copyErr
	}
	if closeErr != nil {
		return "", closeErr
	}

	return filePath, nil
}
