package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/dgyurics/marketplace/types"
)

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conversation *types.Conversation) error
	CreateMessage(ctx context.Context, message *types.Message) error
	GetConversationByID(ctx context.Context, ID string) (types.Conversation, error)
	GetConversationByIDAndUser(ctx context.Context, ID string, userID string) (types.Conversation, error)
	GetConversations(ctx context.Context, userID string) ([]types.Conversation, error)
}

type conversationRepository struct {
	db *sql.DB
}

func NewConversationRepository(db *sql.DB) ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) CreateConversation(ctx context.Context, conversation *types.Conversation) error {
	query := `
		INSERT INTO conversations (id, type, subject, recipient_id)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query,
		conversation.ID,
		conversation.Type,
		conversation.Subject,
		conversation.RecipientID,
	)
	return err
}

func (r *conversationRepository) CreateMessage(ctx context.Context, message *types.Message) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var createdAt time.Time
	query := `
		INSERT INTO messages (id, sender_id, conversation_id, body)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	err = tx.QueryRowContext(ctx, query,
		message.ID,
		message.SenderID,
		message.ConversationID,
		message.Body).Scan(&createdAt)
	if err != nil {
		return err
	}

	query = `
		UPDATE conversations
		SET updated_at = $1
		WHERE id = $2
	`
	_, err = tx.ExecContext(ctx, query, createdAt, message.ConversationID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *conversationRepository) GetConversationByID(ctx context.Context, conversationID string) (types.Conversation, error) {
	var convo types.Conversation

	// Get conversation details
	query := `
		SELECT id, type, subject, recipient_id, recipient_last_read_at, updated_at, created_at
		FROM conversations
		WHERE id = $1
	`
	err := r.db.QueryRowContext(ctx, query, conversationID).Scan(
		&convo.ID, &convo.Type, &convo.Subject,
		&convo.RecipientID, &convo.RecipientLastReadAt, &convo.UpdatedAt, &convo.CreatedAt)
	if err == sql.ErrNoRows {
		return convo, types.ErrNotFound
	}
	if err != nil {
		return convo, err
	}

	// Get messages for conversation
	query = `
		SELECT id, sender_id, conversation_id, body, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return convo, err
	}
	defer rows.Close()

	convo.Messages = []types.Message{}
	for rows.Next() {
		var msg types.Message
		if err = rows.Scan(&msg.ID, &msg.SenderID, &msg.ConversationID, &msg.Body, &msg.CreatedAt); err != nil {
			return convo, err
		}
		convo.Messages = append(convo.Messages, msg)
	}

	return convo, nil
}

func (r *conversationRepository) GetConversationByIDAndUser(ctx context.Context, conversationID string, userID string) (types.Conversation, error) {
	var convo types.Conversation
	// Update recipient last read timestamp + return conversation details
	query := `
		UPDATE conversations
		SET recipient_last_read_at = NOW()
		WHERE id = $1 AND recipient_id = $2
		RETURNING id, type, subject, recipient_id, recipient_last_read_at, updated_at, created_at
	`
	err := r.db.QueryRowContext(ctx, query, conversationID, userID).Scan(
		&convo.ID, &convo.Type, &convo.Subject,
		&convo.RecipientID, &convo.RecipientLastReadAt, &convo.UpdatedAt, &convo.CreatedAt)
	if err == sql.ErrNoRows {
		return convo, types.ErrNotFound
	}
	if err != nil {
		return convo, err
	}

	// Get messages for conversation
	query = `
		SELECT id, sender_id, conversation_id, body, created_at
		FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return convo, err
	}
	defer rows.Close()

	convo.Messages = []types.Message{}
	for rows.Next() {
		var msg types.Message
		if err = rows.Scan(&msg.ID, &msg.SenderID, &msg.ConversationID, &msg.Body, &msg.CreatedAt); err != nil {
			return convo, err
		}
		convo.Messages = append(convo.Messages, msg)
	}

	return convo, nil
}

func (r *conversationRepository) GetConversations(ctx context.Context, userID string) ([]types.Conversation, error) {
	conversations := []types.Conversation{}
	query := `
		SELECT id, type, subject, recipient_id, recipient_last_read_at, updated_at, created_at
		FROM conversations
		WHERE recipient_id = $1 AND is_deleted = FALSE
		ORDER BY updated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return conversations, err
	}
	defer rows.Close()

	for rows.Next() {
		var conversation types.Conversation
		if err := rows.Scan(
			&conversation.ID,
			&conversation.Type,
			&conversation.Subject,
			&conversation.RecipientID,
			&conversation.RecipientLastReadAt,
			&conversation.UpdatedAt,
			&conversation.CreatedAt,
		); err != nil {
			return conversations, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}
