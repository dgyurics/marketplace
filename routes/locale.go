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
	utilities.RespondWithJSON(w, http.StatusOK, utilities.Locale)
}

func (h *LocaleRoutes) RegisterRoutes() {
	h.muxRouter.HandleFunc("/locale", h.GetLocale).Methods("GET")
}
