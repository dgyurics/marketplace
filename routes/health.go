package routes

import (
	"encoding/json"
	"net/http"
)

type HealthRoutes struct {
	router
}

func NewHealthRoutes(router router) *HealthRoutes {
	return &HealthRoutes{
		router: router,
	}
}

func (h *HealthRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/health", h.HealthCheck).Methods(http.MethodGet)
}
func (h *HealthRoutes) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// TODO enhance to check database connection, external services, etc.
	response := map[string]string{"status": "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
