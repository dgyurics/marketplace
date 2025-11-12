package routes

import (
	"net/http"

	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type LocaleRoutes struct {
	router
	config types.LocaleConfig
}

func NewLocaleRoutes(config types.LocaleConfig, router router) *LocaleRoutes {
	return &LocaleRoutes{
		config: config,
		router: router,
	}
}

func (h *LocaleRoutes) GetLocale(w http.ResponseWriter, r *http.Request) {
	data, ok := utilities.LocaleData[h.config.Country]
	if !ok {
		utilities.RespondWithError(w, r, http.StatusNotFound, "locale not found")
		return
	}
	utilities.RespondWithJSON(w, http.StatusOK, data)
}

func (h *LocaleRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/locale", h.GetLocale).Methods("GET")
}
