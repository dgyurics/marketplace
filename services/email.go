package services

import (
	"errors"
	"fmt"
	"sync"
)

type Message struct {
	From        string
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
}

type EmailSender interface {
	Send(message Message) error
}

type MockEmailSender struct {
	SentMessages []Message
	FailOnSend   bool
	mutex        sync.Mutex
}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{
		SentMessages: []Message{},
		FailOnSend:   false,
	}
}

// Send simulates sending an email. It can be configured to fail or succeed.
func (m *MockEmailSender) Send(message Message) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.FailOnSend {
		return errors.New("failed to send email (mock failure)")
	}

	// Store the message in the list of sent messages for later inspection
	m.SentMessages = append(m.SentMessages, message)

	// Simulate sending by printing to the console
	fmt.Printf("Mock Email Sent: From: %s, To: %v, Subject: %s\n", message.From, message.To, message.Subject)
	return nil
}

// GetSentMessages returns a copy of all sent messages for inspection during tests.
func (m *MockEmailSender) GetSentMessages() []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return a copy of the sent messages
	sentMessages := make([]Message, len(m.SentMessages))
	copy(sentMessages, m.SentMessages)
	return sentMessages
}

// ClearSentMessages clears all sent messages (useful between test cases).
func (m *MockEmailSender) ClearSentMessages() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.SentMessages = []Message{}
}

// SetFailOnSend configures whether the mock should fail when sending emails.
func (m *MockEmailSender) SetFailOnSend(fail bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.FailOnSend = fail
}
