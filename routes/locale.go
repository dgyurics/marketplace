package routes

import (
	"net/http"

	"github.com/dgyurics/marketplace/utilities"
)

type LocaleRoutes struct {
	router
}

func NewLocaleRoutes(router router) *LocaleRoutes {
	return &LocaleRoutes{
		router: router,
	}
}

func (h *LocaleRoutes) GetLocale(w http.ResponseWriter, r *http.Request) {
	// handle caching
	etag := `"locale-2025-v1"`
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "public, max-age=2592000") // 1 month

	utilities.RespondWithJSON(w, http.StatusOK, utilities.Locale)
}

func (h *LocaleRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/locale", h.GetLocale).Methods("GET")
}
