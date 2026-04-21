package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
	"github.com/gorilla/mux"
)

type ConversationRoutes struct {
	router
	service services.ConversationService
}

func NewConversationRoutes(service services.ConversationService, router router) *ConversationRoutes {
	return &ConversationRoutes{
		router:  router,
		service: service,
	}
}

func (h *ConversationRoutes) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var conversation types.Conversation
	if err := json.NewDecoder(r.Body).Decode(&conversation); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if conversation.Type == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "type required")
		return
	}
	if conversation.Subject == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "subject required")
		return
	}
	if conversation.RecipientID == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "recipient required")
		return
	}

	if err := h.service.CreateConversation(r.Context(), &conversation); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	u.RespondWithJSON(w, http.StatusCreated, conversation)
}

func (h *ConversationRoutes) GetConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := h.service.GetConversations(r.Context())
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, conversations)
}

func (h *ConversationRoutes) GetConversationAdmin(w http.ResponseWriter, r *http.Request) {
	conversation, err := h.service.GetConversationByID(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, conversation)
}

func (h *ConversationRoutes) GetConversation(w http.ResponseWriter, r *http.Request) {
	conversation, err := h.service.GetConversationByIDAndUser(r.Context(), mux.Vars(r)["id"])
	if err == types.ErrNotFound {
		u.RespondWithError(w, r, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
	u.RespondWithJSON(w, http.StatusOK, conversation)
}

func (h *ConversationRoutes) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var message types.Message
	message.ConversationID = mux.Vars(r)["id"]

	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		u.RespondWithError(w, r, http.StatusBadRequest, "error decoding request payload")
		return
	}

	if message.Body == "" {
		u.RespondWithError(w, r, http.StatusBadRequest, "body required")
		return
	}

	if err := h.service.CreateMessage(r.Context(), &message); err != nil {
		u.RespondWithError(w, r, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *ConversationRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/conversations", h.secure(types.RoleStaff)(h.CreateConversation)).Methods(http.MethodPost)
	h.muxRouter.Handle("/conversations/{id}", h.secure(types.RoleGuest)(h.GetConversation)).Methods(http.MethodGet)
	h.muxRouter.Handle("/conversations/{id}/admin", h.secure(types.RoleStaff)(h.GetConversationAdmin)).Methods(http.MethodGet)
	h.muxRouter.Handle("/conversations/{id}/message", h.secure(types.RoleGuest)(h.CreateMessage)).Methods(http.MethodPost)
	h.muxRouter.Handle("/conversations", h.secure(types.RoleGuest)(h.GetConversations)).Methods(http.MethodGet)
}
