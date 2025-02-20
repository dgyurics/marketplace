package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/dgyurics/marketplace/types"
)

// EmailSender defines a generic interface for sending emails
type EmailSender interface {
	SendEmail(to []string, subject, body string, isHTML bool) error
}

// MailjetSender implements EmailSender using Mailjet's API
type MailjetSender struct {
	Enabled   bool
	APIKey    string
	APISecret string
	FromEmail string
	FromName  string
}

func NewMailjetSender(config types.EmailConfig) *MailjetSender {
	return &MailjetSender{
		Enabled:   config.Enabled,
		APIKey:    config.APIKey,
		APISecret: config.APISecret,
		FromEmail: config.FromEmail,
		FromName:  config.FromName,
	}
}

// MailjetRequest represents the JSON payload for sending an email
type MailjetRequest struct {
	Messages []MailjetMessage `json:"Messages"`
}

// MailjetMessage represents an individual email message
type MailjetMessage struct {
	From         MailjetRecipient   `json:"From"`
	To           []MailjetRecipient `json:"To"`
	Subject      string             `json:"Subject"`
	TextPart     string             `json:"TextPart,omitempty"`
	HTMLPart     string             `json:"HTMLPart,omitempty"`
	CustomID     string             `json:"CustomID"`     // Mark as transactional
	EventPayload string             `json:"EventPayload"` // For tracking purposes
}

// MailjetRecipient represents email sender and recipients
type MailjetRecipient struct {
	Email string `json:"Email"`
	Name  string `json:"Name"`
}

// SendEmail sends an email using Mailjet's API
func (m *MailjetSender) SendEmail(to []string, subject, body string, isHTML bool) error {
	url := "https://api.mailjet.com/v3.1/send"

	// Convert recipient list into Mailjet's format
	var recipients []MailjetRecipient
	for _, email := range to {
		recipients = append(recipients, MailjetRecipient{Email: email, Name: email})
	}

	// Build the email payload
	message := MailjetMessage{
		From:    MailjetRecipient{Email: m.FromEmail, Name: m.FromName},
		To:      recipients,
		Subject: subject,
	}

	if isHTML {
		message.HTMLPart = body
	} else {
		message.TextPart = body
	}

	requestBody, err := json.Marshal(MailjetRequest{Messages: []MailjetMessage{message}})
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(m.APIKey, m.APISecret)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("mailjet error: %s", string(bodyBytes))
	}

	slog.Debug("Email sent successfully", "response", string(bodyBytes))
	return nil
}
