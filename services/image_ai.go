package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log/slog"
	"net/http"
)

const imageEnhancementPrompt = "Keep the product unchanged, including its proportions and size within the image. Do not enlarge or shrink the product. Remove the background. Improve lighting slightly and add a subtle shadow only if needed."

// enhanceImage applies the predefined upload enhancement rules.
func (s *imageService) enhanceImage(ctx context.Context, original []byte) ([]byte, error) {
	contentType := http.DetectContentType(original)
	encodedImage := base64.StdEncoding.EncodeToString(original)
	imageURL := "data:" + contentType + ";base64," + encodedImage

	payload := map[string]any{
		"model":         s.ai.Model,
		"background":    "transparent",
		"output_format": "png",
		"quality":       "medium",
		"size":          "auto",
		"n":             1,
		"images":        []map[string]string{{"image_url": imageURL}},
		"prompt":        imageEnhancementPrompt,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ai.BaseURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.ai.APIKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := s.aiClient.Do(req)
	if err != nil {
		return nil, errors.New("image provider request failed")
	}
	defer res.Body.Close()

	// Read the provider error message.
	if res.StatusCode != http.StatusOK {
		return nil, parseImageProviderError(res.StatusCode, res.Body)
	}

	data, err := readBounded(res.Body, 48<<20)
	if err != nil {
		return nil, err
	}

	var response struct {
		Data []struct {
			Base64 string `json:"b64_json"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, errors.New("invalid image provider response")
	}

	if len(response.Data) != 1 {
		return nil, errors.New("image provider did not return an image")
	}

	decoded, err := base64.StdEncoding.DecodeString(response.Data[0].Base64)
	if err != nil {
		return nil, errors.New("invalid image provider encoding")
	}

	output, err := s.decodeImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("cannot decode enhanced image: %w", err)
	}

	if _, err := png.DecodeConfig(bytes.NewReader(decoded)); err != nil {
		return nil, errors.New("image provider did not return a PNG")
	}

	if !hasTransparentProduct(output) {
		return nil, errors.New("image provider did not return a product on transparency")
	}

	return decoded, nil
}

func hasTransparentProduct(img image.Image) bool {
	transparent, visible := false, false
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			transparent = transparent || c.A == 0
			visible = visible || c.A > 0

			if transparent && visible {
				return true
			}
		}
	}

	return false
}

func readBounded(r io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}

	if int64(len(data)) > limit {
		return nil, errors.New("image response too large")
	}

	return data, nil
}

func (s *imageService) decodeImage(data []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid image data (%s): %w", http.DetectContentType(data), err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > int64(s.maxMegapixels)*1_000_000 {
		return nil, ErrImageResolution
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// ImageProviderError includes the provider status and message.
type ImageProviderError struct {
	status int
	detail string
}

func (e *ImageProviderError) Error() string {
	message := fmt.Sprintf("image provider returned status %d", e.status)
	if e.detail != "" {
		message += ": " + e.detail
	}

	return message
}

func parseImageProviderError(status int, body io.Reader) *ImageProviderError {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	result := &ImageProviderError{status: status}
	raw, err := io.ReadAll(io.LimitReader(body, 16<<10))
	slog.Warn("Image provider request failed", "status", status, "body", string(raw), "read_error", err)
	if err := json.Unmarshal(raw, &response); err == nil {
		result.detail = response.Error.Message
	}

	return result
}
