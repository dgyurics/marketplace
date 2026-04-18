package services

import (
	"context"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
)

type ConversationService interface {
	CreateConversation(ctx context.Context, conversation *types.Conversation) error
	CreateMessage(ctx context.Context, conversationID string, message *types.Message) error
	GetConversation(ctx context.Context, ID string) (types.Conversation, error)
	GetConversations(ctx context.Context) ([]types.Conversation, error)
}

type conversationService struct {
	repo repositories.ConversationRepository
}

func NewConversationService(repo repositories.ConversationRepository) ConversationService {
	return &conversationService{repo: repo}
}

func (s *conversationService) CreateConversation(ctx context.Context, conversation *types.Conversation) error {
	return nil
}

func (s *conversationService) CreateMessage(ctx context.Context, conversationID string, message *types.Message) error {
	return nil
}

func (s *conversationService) GetConversation(ctx context.Context, ID string) (types.Conversation, error) {
	var conversation types.Conversation
	return conversation, nil
}

func (s *conversationService) GetConversations(ctx context.Context) ([]types.Conversation, error) {
	return nil, nil
}
