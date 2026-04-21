package types

import "time"

type ConversationType string

// Conversation types:
// - notification: System/admin initiated one-way communication to notify users of events
// - support: Admin response to customer tickets, allows customer to reply back and forth
const (
	Support      ConversationType = "support"
	Notification ConversationType = "notification"
)

type Conversation struct {
	ID                  string           `json:"id"`
	Type                ConversationType `json:"type"`
	Subject             string           `json:"subject"`
	RecipientID         string           `json:"recipient_id"`
	RecipientLastReadAt time.Time        `json:"recipient_last_read_at"`
	Messages            []Message        `json:"messages"`
	IsDeleted           bool             `json:"-"`
	UpdatedAt           time.Time        `json:"updated_at"`
	CreatedAt           time.Time        `json:"created_at"`
}

type Message struct {
	ID             string    `json:"id"`
	SenderID       string    `json:"sender_id"`
	ConversationID string    `json:"conversation_id"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"created_at"`
}
