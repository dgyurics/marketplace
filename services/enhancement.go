package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

const (
	maxImageBytes  = 32 << 20
	maxImagePixels = 24_000_000
)

var ErrInvalidOriginalImage = errors.New("original image cannot be decoded")

const DefaultImagePrompt = "Keep the product unchanged. Add a subtle shadow if needed. Remove any background, if one exists. Return a transparent PNG."

var (
	ErrImageContextInvalid = errors.New("image instructions must contain at most 1000 characters")
	ErrImageAIDisabled     = errors.New("AI image improvement is not configured")
	ErrImagePreviewStale   = errors.New("image changed or preview expired; generate a new preview")
)

type EnhancementService interface {
	Generate(ctx context.Context, imageID, prompt string) (Enhancement, error)
	Accept(ctx context.Context, imageID, previewID string) error
	Discard(ctx context.Context, imageID, previewID string) error
}

type enhancementService struct {
	repo   repositories.ImageRepository
	config types.ImageConfig
	ai     types.AIModelConfig
	client *http.Client
}

// NewEnhancementService initializes preview processing.
func NewEnhancementService(
	repo repositories.ImageRepository,
	config types.ImageConfig,
	ai types.AIModelConfig) EnhancementService {
	service := &enhancementService{
		repo:   repo,
		config: config,
		ai:     ai,
		client: &http.Client{
			Timeout: 150 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}

	cleanupExpiredAIArtifacts(config.ImageUploadPath)

	return service
}

type Enhancement struct {
	ID        string `json:"id"`
	BeforeURL string `json:"before_url"`
	AfterURL  string `json:"after_url"`
}

// previewFilename binds a preview to the image and source it was generated from.
func previewFilename(img types.Image, token string) (string, error) {
	data, err := hex.DecodeString(token)
	if err != nil || len(data) != 16 {
		return "", ErrImagePreviewStale
	}
	digest := sha256.Sum256([]byte(img.ID + ":" + img.Source))
	return fmt.Sprintf(".ai-preview-%x-%s.png", digest, token), nil
}

// randomImageToken creates a preview identifier.
func randomImageToken() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	return hex.EncodeToString(b[:]), nil
}

// Generate waits for the provider and saves the completed preview.
func (s *enhancementService) Generate(ctx context.Context, id, productContext string) (Enhancement, error) {
	if !s.ai.Enabled {
		return Enhancement{}, ErrImageAIDisabled
	}

	if strings.TrimSpace(productContext) == "" {
		productContext = DefaultImagePrompt
	}

	if !utf8.ValidString(productContext) || utf8.RuneCountInString(productContext) > 1000 {
		return Enhancement{}, ErrImageContextInvalid
	}

	cleanupExpiredAIArtifacts(s.config.ImageUploadPath)

	img, err := s.repo.GetImage(ctx, id)
	if err != nil {
		return Enhancement{}, err
	}

	data, err := s.improveImage(ctx, img, productContext)
	if err != nil {
		return Enhancement{}, err
	}
	if err := ctx.Err(); err != nil {
		return Enhancement{}, err
	}
	token, err := randomImageToken()
	if err != nil {
		return Enhancement{}, err
	}
	name, err := previewFilename(img, token)
	if err != nil {
		return Enhancement{}, err
	}
	if err := writeNewFile(filepath.Join(s.config.ImageUploadPath, img.ProductID, name), data); err != nil {
		return Enhancement{}, err
	}
	return Enhancement{ID: token, BeforeURL: img.URL, AfterURL: generateImageURL(s.config, img.ProductID, name, img.Type)}, nil
}

// improveImage reads and enhances the saved image.
func (s *enhancementService) improveImage(ctx context.Context, img types.Image, productContext string) ([]byte, error) {
	originalPath := filepath.Join(s.config.ImageUploadPath, img.ProductID, filepath.Base(img.Source))
	f, err := os.Open(originalPath)
	if err != nil {
		return nil, err
	}

	original, err := readBounded(f, maxImageBytes)
	f.Close()
	if err != nil {
		return nil, err
	}

	return s.enhanceImage(ctx, original, productContext)
}

// Accept replaces the saved image and its variants.
func (s *enhancementService) Accept(ctx context.Context, id, token string) error {
	img, err := s.repo.GetImage(ctx, id)
	if err != nil {
		return err
	}
	filename, err := previewFilename(img, token)
	if err != nil {
		return err
	}
	variants, err := s.repo.GetSourceImages(ctx, img.ProductID, img.Source)
	if err != nil {
		return err
	}

	// Accepted files have immutable, non-preview URLs; stale caches cannot hide the change.
	fileID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}

	finalName := fileID + ".png"
	dir := filepath.Join(s.config.ImageUploadPath, img.ProductID)
	err = os.Rename(filepath.Join(dir, filename), filepath.Join(dir, finalName))
	if os.IsNotExist(err) {
		return ErrImagePreviewStale
	}
	if err != nil {
		return err
	}

	for i := range variants {
		variants[i].Source = finalName
		variants[i].URL = generateImageURL(s.config, img.ProductID, finalName, variants[i].Type)
	}

	if err := s.repo.ReplaceSource(ctx, img.ProductID, img.Source, variants); err != nil {
		_ = os.Rename(filepath.Join(dir, finalName), filepath.Join(dir, filename))
		return fmt.Errorf("%w: %v", ErrImagePreviewStale, err)
	}

	_ = os.Remove(filepath.Join(dir, img.Source))
	return nil
}

// Discard removes a completed preview.
func (s *enhancementService) Discard(ctx context.Context, id, token string) error {
	img, err := s.repo.GetImage(ctx, id)
	if err != nil {
		return err
	}
	filename, err := previewFilename(img, token)
	if err != nil {
		return err
	}
	err = os.Remove(filepath.Join(s.config.ImageUploadPath, img.ProductID, filename))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// cleanupExpiredAIArtifacts removes abandoned preview files.
func cleanupExpiredAIArtifacts(imgDir string) {
	products, err := os.ReadDir(imgDir)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-15 * time.Minute)
	for _, product := range products {
		if !product.IsDir() {
			continue
		}

		cleanupExpiredPreviews(filepath.Join(imgDir, product.Name()), cutoff)
	}
}

func cleanupExpiredPreviews(dir string, cutoff time.Time) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() || (!strings.HasPrefix(file.Name(), ".ai-preview-") && !strings.HasPrefix(file.Name(), ".ai-work-")) {
			continue
		}
		info, err := file.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(dir, file.Name()))
	}
}

// writeNewFile saves a new file without overwriting.
func writeNewFile(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	closeErr := f.Close()
	if err == nil {
		err = closeErr
	}

	if err != nil {
		_ = os.Remove(path)
	}

	return err
}

// enhanceImage sends the image and prompt to the AI provider.
func (s *enhancementService) enhanceImage(ctx context.Context, original []byte, prompt string) ([]byte, error) {
	if _, err := decodeImage(original); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOriginalImage, err)
	}

	contentType := http.DetectContentType(original)
	encodedImage := base64.StdEncoding.EncodeToString(original)
	imageURL := "data:" + contentType + ";base64," + encodedImage

	// OpenAI Images API params; support varies by provider at BaseURL
	payload := map[string]any{
		"model":         s.ai.Model,    // image model
		"background":    "transparent", // needs png/webp; verified by hasTransparentProduct
		"output_format": "png",         // png | jpeg | webp
		"quality":       "medium",      // low | medium | high | auto
		"size":          "auto",        // model picks dimensions
		"n":             1,             // image count; exactly one required
		"images": []map[string]string{ // JSON variant of /v1/images/edits
			{"image_url": imageURL}, // base64 data URL of the original
		},
		"prompt": prompt, // user instructions
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

	res, err := s.client.Do(req)
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

	output, err := decodeImage(decoded)
	if err != nil {
		return nil, fmt.Errorf("cannot decode OpenAI output: %w", err)
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

func decodeImage(data []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("invalid image data (%s): %w", http.DetectContentType(data), err)
	}

	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > maxImagePixels {
		return nil, errors.New("image dimensions too large")
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
	if err := json.NewDecoder(io.LimitReader(body, 16<<10)).Decode(&response); err == nil {
		result.detail = response.Error.Message
	}

	return result
}
