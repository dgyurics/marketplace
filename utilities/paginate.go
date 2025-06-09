package utilities

import (
	"net/http"
	"strconv"
)

type PaginationParams struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func ParsePaginationParams(r *http.Request, defaultPage, defaultLimit int) PaginationParams {
	// Default values
	page := defaultPage
	limit := defaultLimit

	// Parse query parameters
	query := r.URL.Query()
	if p, err := strconv.Atoi(query.Get("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(query.Get("limit")); err == nil && l > 0 {
		limit = l
	}

	if limit > 100 {
		limit = 100 // Cap the limit to a maximum of 100
	}

	return PaginationParams{Page: page, Limit: limit}
}
