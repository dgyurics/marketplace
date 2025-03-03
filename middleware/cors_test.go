package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgyurics/marketplace/types"
	"github.com/stretchr/testify/require"
)

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}

	middleware := CORSMiddleware(config)

	// Create request with allowed origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:8081")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Validate response headers
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "http://localhost:8081", rr.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
	require.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}

	middleware := CORSMiddleware(config)

	// Create request with a different origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://unauthorized.com")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Expect Forbidden response
	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCORSMiddleware_ValidOptionsRequest(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}

	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:8081")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Expect a 200 OK for valid OPTIONS request
	require.Equal(t, http.StatusOK, rr.Code)
}

func TestCORSMiddleware_DisallowedOptionsRequest(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}

	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://unauthorized.com")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Expect 403 Forbidden for preflight request from disallowed origin
	require.Equal(t, http.StatusForbidden, rr.Code)
}

func TestCORSMiddleware_AllowCredentials(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"http://localhost:8081"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}

	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:8081")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Validate CORS response headers
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "http://localhost:8081", rr.Header().Get("Access-Control-Allow-Origin"))
	require.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_WildcardOrigin(t *testing.T) {
	config := types.CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
	}

	middleware := CORSMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://random.com")

	rr := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rr, req)

	// Expect wildcard "*" for allowed origin
	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
}
