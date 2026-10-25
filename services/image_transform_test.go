package services_test

import (
	"context"
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
				Enabled: true,
				BaseURL: server.URL,
				Model:   "test-model",
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
