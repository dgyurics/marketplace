package services

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

type Template string

const (
	PasswordReset              Template = "password_reset.html"
	OrderConfirmation          Template = "order_confirmation.html"
	OrderNotification          Template = "order_notification.html"
	EmailVerification          Template = "email_verification.html"
	PurchaseIntentUpdate       Template = "purchase_intent_update.html"
	PurchaseIntentNotification Template = "purchase_intent_notification.html"
)

// TemplateService defines methods for managing HTML templates.
type TemplateService interface {
	Render(w io.Writer, name Template, data interface{}) error
	RenderToString(name Template, data interface{}) (string, error)
}

type templateService struct {
	templates *template.Template
}

// NewTemplateService initializes and loads all templates from the given directory.
func NewTemplateService(templateDir string) TemplateService {
	tmpl, err := loadTemplates(templateDir)
	if err != nil {
		slog.Error("Failed to load templates", "error", err, "templateDir", templateDir)
		os.Exit(1)
	}
	return &templateService{templates: tmpl}
}

// loadTemplates parses all templates in the specified directory.
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

// Render renders a template by name and writes the output to an io.Writer.
func (s *templateService) Render(w io.Writer, name Template, data interface{}) error {
	return s.templates.ExecuteTemplate(w, string(name), data)
}

// RenderToString renders a template to a string.
func (s *templateService) RenderToString(name Template, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := s.Render(&buf, name, data)
	return buf.String(), err
}
