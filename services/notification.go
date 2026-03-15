package services

import (
	"log/slog"

	"github.com/dgyurics/marketplace/types"
)

type NotificationService interface {
	BaseURL() string
	SendEmail(to, subject string, templateName Template, data interface{}) error
}

type notificationService struct {
	emailService    EmailService
	templateService TemplateService
	baseURL         string
}

func NewNotificationService(emailService EmailService, templateService TemplateService, baseURL string) NotificationService {
	return &notificationService{
		emailService:    emailService,
		templateService: templateService,
		baseURL:         baseURL,
	}
}

func (s *notificationService) BaseURL() string {
	return s.baseURL
}

func (s *notificationService) SendEmail(to, subject string, template Template, data interface{}) error {
	htmlBody, err := s.templateService.RenderToString(template, data)
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
