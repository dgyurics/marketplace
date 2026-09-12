package routes

import (
	"net/http"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ConfigRoutes struct {
	config types.AppMetadata
	router
}

func NewConfigRoutes(config types.AppMetadata, router router) *ConfigRoutes {
	return &ConfigRoutes{
		config: config,
		router: router,
	}
}

func (h *ConfigRoutes) GetConfig(w http.ResponseWriter, r *http.Request) {
	// handle caching
	etag := `"app-config-2026-v1"`
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=2592000") // 1 month

	utilities.RespondWithJSON(w, http.StatusOK, h.config)
}

func (h *ConfigRoutes) RegisterRoutes() {
	h.mux.HandleFunc("GET /config", h.GetConfig)
}
