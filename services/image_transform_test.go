package services_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageTransformService_Disabled(t *testing.T) {
	// No client or source file is needed when transformations are disabled.
	service := services.NewImageTransformService(nil, types.ImageTransformConfig{})
	path := filepath.Join(t.TempDir(), "image.png")

	result, err := service.Transform(context.Background(), path, "image.png")

	require.NoError(t, err)
	assert.Equal(t, path, result)
}

func TestImageTransformService_Transform(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
	}{
		{name: "success", status: http.StatusOK},
		{name: "provider error", status: http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "image.png")
			require.NoError(t, os.WriteFile(path, []byte("original image"), 0600))

			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/remove", r.URL.Path)
				file, header, err := r.FormFile("file")
				if !assert.NoError(t, err) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				defer file.Close()
				defer r.MultipartForm.RemoveAll()
				contents, err := io.ReadAll(file)
				assert.NoError(t, err)
				assert.Equal(t, "image.png", header.Filename)
				assert.Equal(t, "original image", string(contents))
				assert.Equal(t, "test-model", r.FormValue("model"))
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, "transformed image")
			}))
			defer server.Close()

			service := services.NewImageTransformService(server.Client(), types.ImageTransformConfig{
				Enabled:  true,
				Provider: types.ImageTransformProviderRembg,
				BaseURL:  server.URL,
				Model:    "test-model",
			})
			result, err := service.Transform(context.Background(), path, "image.png")
			assert.True(t, called)
			if tc.status == http.StatusOK {
				require.NoError(t, err)
				assert.Equal(t, path, result)
			} else {
				require.Error(t, err)
			}

			contents, err := os.ReadFile(path)
			require.NoError(t, err)
			if tc.status == http.StatusOK {
				assert.Equal(t, "transformed image", string(contents))
			} else {
				assert.Equal(t, "original image", string(contents))
			}
		})
	}
}

func TestImageTransformService_UnsupportedProvider(t *testing.T) {
	service := services.NewImageTransformService(nil, types.ImageTransformConfig{
		Enabled:  true,
		Provider: "unknown",
	})

	_, err := service.Transform(context.Background(), "image.png", "image.png")

	require.ErrorContains(t, err, "unsupported image transform provider")
}

func TestImageTransformService_OpenAI(t *testing.T) {
	var original, output bytes.Buffer
	require.NoError(t, png.Encode(&original, image.NewNRGBA(image.Rect(0, 0, 1, 1))))
	require.NoError(t, png.Encode(&output, image.NewNRGBA(image.Rect(0, 0, 2, 2))))
	encoded := base64.StdEncoding.EncodeToString(output.Bytes())
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		wantError string
	}{
		{name: "success", status: http.StatusOK, body: `{"data":[{"b64_json":"` + encoded + `"}]}`},
		{name: "provider error", status: http.StatusUnauthorized, body: `{"error":{"message":"Invalid API key"}}`, wantError: "401 Unauthorized: Invalid API key"},
		{name: "invalid response", status: http.StatusOK, body: `not json`, wantError: "decode openai response"},
		{name: "missing image", status: http.StatusOK, body: `{"data":[]}`, wantError: "did not return an image"},
		{name: "invalid encoding", status: http.StatusOK, body: `{"data":[{"b64_json":"!"}]}`, wantError: "decode openai image"},
		{name: "invalid image", status: http.StatusOK, body: `{"data":[{"b64_json":"aGVsbG8="}]}`, wantError: "valid PNG"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "image.png")
			require.NoError(t, os.WriteFile(path, original.Bytes(), 0600))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/v1/images/edits", r.URL.Path)
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				var payload struct {
					Model        string `json:"model"`
					Prompt       string `json:"prompt"`
					Background   string `json:"background"`
					OutputFormat string `json:"output_format"`
					Images       []struct {
						URL string `json:"image_url"`
					} `json:"images"`
				}
				if assert.NoError(t, json.NewDecoder(r.Body).Decode(&payload)) {
					assert.Equal(t, "test-model", payload.Model)
					assert.NotEmpty(t, payload.Prompt)
					assert.Equal(t, "transparent", payload.Background)
					assert.Equal(t, "png", payload.OutputFormat)
					if assert.Len(t, payload.Images, 1) {
						assert.Equal(t, "data:image/png;base64,"+base64.StdEncoding.EncodeToString(original.Bytes()), payload.Images[0].URL)
					}
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			defer server.Close()
			service := services.NewImageTransformService(server.Client(), types.ImageTransformConfig{
				Enabled:  true,
				Provider: types.ImageTransformProviderOpenAI,
				BaseURL:  server.URL + "/v1/images/edits",
				Model:    "test-model",
				APIKey:   "test-key",
			})
			result, err := service.Transform(context.Background(), path, "image.png")
			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, path, result)
			} else {
				require.ErrorContains(t, err, tc.wantError)
			}
			contents, err := os.ReadFile(path)
			require.NoError(t, err)
			if tc.wantError == "" {
				assert.Equal(t, output.Bytes(), contents)
			} else {
				assert.Equal(t, original.Bytes(), contents)
			}
		})
	}
}
