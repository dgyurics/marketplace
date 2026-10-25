package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
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
	switch s.config.Provider {
	case types.ImageTransformProviderRembg:
		return s.removeBackground(ctx, filePath, filename)
	case types.ImageTransformProviderOpenAI:
		return s.transformOpenAI(ctx, filePath)
	default:
		return "", fmt.Errorf("unsupported image transform provider: %q", s.config.Provider)
	}
}

const imageEnhancementPrompt = "Keep the product unchanged, including its proportions and size within the image. Do not enlarge or shrink the product. Remove the background. Improve lighting slightly and add a subtle shadow only if needed."

// transformOpenAI requests an enhanced PNG, then overwrites the stored image.
// BaseURL is the complete image-edit endpoint, e.g. https://api.openai.com/v1/images/edits.
func (s *imageTransformService) transformOpenAI(ctx context.Context, filePath string) (string, error) {
	if s.config.APIKey == "" || s.config.Model == "" {
		return "", fmt.Errorf("openai image transformation requires an API key and model")
	}

	original, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}
	imageURL := "data:" + http.DetectContentType(original) + ";base64," + base64.StdEncoding.EncodeToString(original)
	payload := map[string]any{
		"model":         s.config.Model,
		"images":        []map[string]string{{"image_url": imageURL}},
		"prompt":        imageEnhancementPrompt,
		"background":    "transparent",
		"output_format": "png",
		"quality":       "medium",
		"size":          "auto",
		"n":             1,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to encode openai request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.BaseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to openai: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var failure struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(io.LimitReader(res.Body, 16<<10)).Decode(&failure); err == nil && failure.Error.Message != "" {
			return "", fmt.Errorf("openai image transformation failed: %s: %s", res.Status, failure.Error.Message)
		}
		return "", fmt.Errorf("openai image transformation failed: %s", res.Status)
	}

	var response struct {
		Data []struct {
			Base64 string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 48<<20)).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode openai response: %w", err)
	}
	if len(response.Data) != 1 || response.Data[0].Base64 == "" {
		return "", fmt.Errorf("openai did not return an image")
	}
	output, err := base64.StdEncoding.DecodeString(response.Data[0].Base64)
	if err != nil {
		return "", fmt.Errorf("failed to decode openai image: %w", err)
	}
	if _, err := png.DecodeConfig(bytes.NewReader(output)); err != nil {
		return "", fmt.Errorf("openai did not return a valid PNG: %w", err)
	}
	if err := os.WriteFile(filePath, output, 0600); err != nil {
		return "", fmt.Errorf("failed to save transformed image: %w", err)
	}
	return filePath, nil
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
