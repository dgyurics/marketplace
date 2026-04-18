package routes

import (
	"encoding/json"
	"net/http"

	"github.com/dgyurics/marketplace/services"
	"github.com/dgyurics/marketplace/types"
	u "github.com/dgyurics/marketplace/utilities"
)

type ConversationRoutes struct {
	router
	service services.ConversationService
}

func NewConversationRoutes(router router) *ConversationRoutes {
	return &ConversationRoutes{
		router: router,
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
	// retrieves slice of conversations without populating messages
}

func (h *ConversationRoutes) GetConversation(w http.ResponseWriter, r *http.Request) {
	// retrieves a single conversation, populating all associated messages
}

func (h *ConversationRoutes) CreateMessage(w http.ResponseWriter, r *http.Request) {
	// conversationID := mux.Vars(r)["id"]
	// h.service.CreateMessage()
}

func (h *ConversationRoutes) RegisterRoutes() {
	h.muxRouter.Handle("/conversations", h.secure(types.RoleAdmin)(h.CreateConversation)).Methods(http.MethodPost)
	h.muxRouter.Handle("/conversations", h.secure(types.RoleAdmin)(h.GetConversations)).Methods(http.MethodGet)
	h.muxRouter.Handle("/conversations/{id}", h.secure(types.RoleAdmin)(h.GetConversation)).Methods(http.MethodGet)
	h.muxRouter.Handle("/conversations/{id}/message", h.secure(types.RoleGuest)(h.CreateMessage)).Methods(http.MethodPost)
}
