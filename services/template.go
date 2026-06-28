package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"os"
	"path/filepath"
)

const templatesDir = "./utilities/templates"

const (
	SubjectPasswordReset string = "password reset"
	SubjectEmailVerify   string = "verify your email"
	SubjectOrderConf     string = "order confirmation"
	SubjectOrderUpdate   string = "order update"
	SubjectOrderRecv     string = "new order received"
	SubjectOfferConf     string = "offer confirmation"
	SubjectOfferUpdate   string = "offer update"
	SubjectOfferRecv     string = "new offer received"
)

// HtmlTemplate identifies a template file by name.
type HtmlTemplate string

// Email templates (sent via SMTP)
const (
	EmailPasswordReset HtmlTemplate = "email_password_reset.html"
	EmailVerification  HtmlTemplate = "email_verification.html"
	EmailOrderConf     HtmlTemplate = "email_order_confirmation.html"
	EmailOfferConf     HtmlTemplate = "email_offer_confirmation.html"
)

// Notification templates (rendered in the user inbox)
const (
	NotifyOrderConf   HtmlTemplate = "notify_order_confirmation.html"
	NotifyOrderUpdate HtmlTemplate = "notify_order_update.html" // Not yet implemented
	NotifyOrderRecv   HtmlTemplate = "notify_order_received.html"
	NotifyOfferUpdate HtmlTemplate = "notify_offer_update.html"
	NotifyOfferConf   HtmlTemplate = "notify_offer_confirmation.html"
	NotifyOfferRecv   HtmlTemplate = "notify_offer_received.html"
)

// TemplateService renders named HTML templates with the provided data.
type TemplateService interface {
	RenderHtmlToString(name HtmlTemplate, data interface{}) (string, error)
}

type templateService struct {
	templates *template.Template
}

// NewTemplateService loads all HTML templates at startup.
func NewTemplateService() TemplateService {
	templates, err := loadTemplates(templatesDir)
	if err != nil {
		slog.Error("Failed to load templates", "error", err, "directory", templatesDir)
		os.Exit(1)
	}
	return &templateService{templates}
}

// loadTemplates parses all .html files in the given directory into a single template set.
func loadTemplates(templateDir string) (*template.Template, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no templates found in directory: %s", templateDir)
	}
	return template.ParseFiles(files...)
}

// RenderHtmlToString executes the named template with the given data and returns the result.
func (s *templateService) RenderHtmlToString(name HtmlTemplate, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := s.templates.ExecuteTemplate(&buf, string(name), data)
	return buf.String(), err
}
