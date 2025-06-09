package utilities

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePaginationParams_DefaultValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, defaultPage, params.Page)
	assert.Equal(t, defaultLimit, params.Limit)
}

func TestParsePaginationParams_ValidValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=5&limit=20", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, 5, params.Page)
	assert.Equal(t, 20, params.Limit)
}

func TestParsePaginationParams_InvalidPageValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=abc&limit=15", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, defaultPage, params.Page) // Should fall back to default
	assert.Equal(t, 15, params.Limit)         // Valid limit should be applied
}

func TestParsePaginationParams_InvalidLimitValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=2&limit=-5", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, 2, params.Page)             // Valid page should be applied
	assert.Equal(t, defaultLimit, params.Limit) // Should fall back to default
}

func TestParsePaginationParams_ZeroValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=0&limit=0", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, defaultPage, params.Page)   // Should fall back to default
	assert.Equal(t, defaultLimit, params.Limit) // Should fall back to default
}

func TestParsePaginationParams_ExceedingLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?page=1&limit=150", nil)

	defaultPage := 1
	defaultLimit := 10
	params := ParsePaginationParams(req, defaultPage, defaultLimit)

	assert.Equal(t, 1, params.Page)    // Valid page should be applied
	assert.Equal(t, 100, params.Limit) // Limit should be capped at 100
}
