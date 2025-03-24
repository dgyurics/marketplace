package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a test handler
func testHandler(w http.ResponseWriter, r *http.Request) {
	// Try reading the body to trigger the middleware limit
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func TestLimitBodySizeMiddleware_WithinLimit(t *testing.T) {
	handler := LimitBodySizeMiddleware(http.HandlerFunc(testHandler))

	// Create a request with a body size within the limit
	body := strings.Repeat("a", int(MaxBodyBytes-1))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, body, string(responseBody))
}

func TestLimitBodySizeMiddleware_AtLimit(t *testing.T) {
	handler := LimitBodySizeMiddleware(http.HandlerFunc(testHandler))

	// Create a request with a body exactly at the limit
	body := strings.Repeat("a", int(MaxBodyBytes))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	responseBody, _ := io.ReadAll(res.Body)
	assert.Equal(t, body, string(responseBody))
}

func TestLimitBodySizeMiddleware_ExceedsLimit(t *testing.T) {
	handler := LimitBodySizeMiddleware(http.HandlerFunc(testHandler))

	// Create a request with a body exceeding the limit
	body := strings.Repeat("a", int(MaxBodyBytes+1))
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)

	responseBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(responseBody), "Request body too large")
}
