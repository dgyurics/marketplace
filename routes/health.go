package routes

import (
	"net/http"

	u "github.com/dgyurics/marketplace/utilities"
)

type HealthRoutes struct {
	router
}

func NewHealthRoutes(router router) *HealthRoutes {
	return &HealthRoutes{
		router: router,
	}
}

func (h *HealthRoutes) HealthCheck(w http.ResponseWriter, r *http.Request) {
	u.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthRoutes) RegisterRoutes() {
	h.mux.Handle("GET /health", http.HandlerFunc(h.HealthCheck))
}
