package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

type ConversationService interface {
	CreateConversation(ctx context.Context, conversation *types.Conversation) error
	CreateMessage(ctx context.Context, message *types.Message) error
	GetConversationByID(ctx context.Context, conversationID string) (types.Conversation, error)
	GetConversationByIDAndUser(ctx context.Context, conversationID string) (types.Conversation, error)
	GetConversations(ctx context.Context) ([]types.Conversation, error)
}

type conversationService struct {
	repo repositories.ConversationRepository
}

func NewConversationService(repo repositories.ConversationRepository) ConversationService {
	return &conversationService{repo: repo}
}

func (s *conversationService) CreateConversation(ctx context.Context, conversation *types.Conversation) error {
	convID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	conversation.ID = convID
	return s.repo.CreateConversation(ctx, conversation)
}

func (s *conversationService) CreateMessage(ctx context.Context, message *types.Message) error {
	message.SenderID = getUserID(ctx)
	messageID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}
	message.ID = messageID
	return s.repo.CreateMessage(ctx, message)
}

func (s *conversationService) GetConversationByID(ctx context.Context, conversationID string) (types.Conversation, error) {
	return s.repo.GetConversationByID(ctx, conversationID)
}

func (s *conversationService) GetConversationByIDAndUser(ctx context.Context, conversationID string) (types.Conversation, error) {
	return s.repo.GetConversationByIDAndUser(ctx, conversationID, getUserID(ctx))
}

func (s *conversationService) GetConversations(ctx context.Context) ([]types.Conversation, error) {
	return s.repo.GetConversations(ctx, getUserID(ctx))
}
