package repositories

import (
	"context"
	"database/sql"

	"github.com/dgyurics/marketplace/types"
)

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conversation *types.Conversation) error
	CreateMessage(ctx context.Context, conversationID string, message *types.Message) error
	GetConversation(ctx context.Context, ID string) (types.Conversation, error)
	GetConversations(ctx context.Context) ([]types.Conversation, error)
}

type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) CreateConversation(ctx context.Context, conversation *types.Conversation) error {
	return nil
}

func (r *conversationRepository) CreateMessage(ctx context.Context, conversationID string, message *types.Message) error {
	return nil
}

func (r *conversationRepository) GetConversation(ctx context.Context, ID string) (types.Conversation, error) {
	var conversation types.Conversation
	return conversation, nil
}

func (r *conversationRepository) GetConversations(ctx context.Context) ([]types.Conversation, error) {
	return nil, nil
}
