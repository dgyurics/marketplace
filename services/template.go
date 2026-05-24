package services

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"log/slog"
	"os"
	"path/filepath"
	texttemplate "text/template"
)

type HtmlTemplate string

const (
	PasswordReset     HtmlTemplate = "password_reset.html"
	OrderConfirmation HtmlTemplate = "order_confirmation.html"
	OrderNotification HtmlTemplate = "order_notification.html"
	EmailVerification HtmlTemplate = "email_verification.html"
	OfferUpdate       HtmlTemplate = "offer_update.html"
	OfferNotification HtmlTemplate = "offer_notification.html"
)

type TextTemplate string

const (
	// these go to admins when an order/offer is received
	OrderNotificationMessage TextTemplate = "order_notification.txt"
	OfferNotificationMessage TextTemplate = "offer_notification.txt"

	// these go to users when an order/offer is created
	OrderConfirmationMessage TextTemplate = "order_confirmation.txt"
	OfferConfirmationMessage TextTemplate = "offer_confirmation.txt"

	// these go to users when an order/offer status changes
	OrderUpdateMessage TextTemplate = "order_update.txt"
	OfferUpdateMessage TextTemplate = "offer_update.txt"
)

// TemplateService defines the interface for rendering templates.
type TemplateService interface {
	RenderHtmlToString(name HtmlTemplate, data interface{}) (string, error)
	RenderTextToString(name TextTemplate, data interface{}) (string, error)
}

type templateService struct {
	templates     *htmltemplate.Template
	textTemplates *texttemplate.Template
}

// NewTemplateService initializes and loads all templates from the given directory.
func NewTemplateService(templateDir string) TemplateService {
	tmpl, err := loadTemplates(templateDir)
	if err != nil {
		slog.Error("Failed to load templates", "error", err, "templateDir", templateDir)
		os.Exit(1)
	}
	txtTmpl, err := loadTextTemplates(templateDir)
	if err != nil {
		slog.Error("Failed to load templates", "error", err, "templateDir", templateDir)
		os.Exit(1)
	}
	return &templateService{templates: tmpl, textTemplates: txtTmpl}
}

// loadTemplates parses all html templates in the specified directory.
func loadTemplates(templateDir string) (*htmltemplate.Template, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no templates found in directory: %s", templateDir)
	}

	return htmltemplate.ParseFiles(files...)
}

// loadTextTemplates parses all text templates in the specified directory.
func loadTextTemplates(templateDir string) (*texttemplate.Template, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.txt"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no templates found in directory: %s", templateDir)
	}
	return texttemplate.ParseFiles(files...)
}

// RenderHtmlToString renders an html template to a string.
func (s *templateService) RenderHtmlToString(name HtmlTemplate, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := s.templates.ExecuteTemplate(&buf, string(name), data)
	return buf.String(), err
}

// RenderTextToString renders a text template to a string.
func (s *templateService) RenderTextToString(name TextTemplate, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := s.textTemplates.ExecuteTemplate(&buf, string(name), data)
	return buf.String(), err
}
