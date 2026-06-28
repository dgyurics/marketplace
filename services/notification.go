package services

import (
	"context"
	"log/slog"

	"github.com/dgyurics/marketplace/types"
)

type NotificationService interface {
	BaseURL() string
	Notify(to, subject string, template HtmlTemplate, data interface{}) error
	SendEmail(to, subject string, tempalte HtmlTemplate, data interface{}) error
}

type notificationService struct {
	emailService        EmailService
	templateService     TemplateService
	conversationService ConversationService
	baseURL             string
}

func NewNotificationService(emailService EmailService, templateService TemplateService, conversationService ConversationService, baseURL string) NotificationService {
	return &notificationService{
		emailService:        emailService,
		templateService:     templateService,
		conversationService: conversationService,
		baseURL:             baseURL,
	}
}

func (s *notificationService) BaseURL() string {
	return s.baseURL
}

func (s *notificationService) SendEmail(to, subject string, template HtmlTemplate, data interface{}) error {
	htmlBody, err := s.templateService.RenderHtmlToString(template, data)
	if err != nil {
		slog.Error("Error loading email template: ", "template", template, "error", err)
		return err
	}
	email := &types.Email{
		To:      []string{to},
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	}
	return s.emailService.Send(email)
}

func (s *notificationService) Notify(to, subject string, template HtmlTemplate, data interface{}) error {
	ctx := systemContext()

	conv := &types.Conversation{
		RecipientID: to,
		Type:        types.Notification,
		Subject:     subject,
	}
	if err := s.conversationService.CreateConversation(ctx, conv); err != nil {
		return err
	}

	body, err := s.templateService.RenderHtmlToString(template, data)
	if err != nil {
		return err
	}

	msg := &types.Message{
		ConversationID: conv.ID,
		Body:           body,
	}

	return s.conversationService.CreateMessage(ctx, msg)
}

const systemUserID = "1"

func systemContext() context.Context {
	return context.WithValue(context.Background(), UserKey, &types.User{ID: systemUserID, Role: "system"})
}
