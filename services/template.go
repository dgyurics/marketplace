package services

import (
	"bytes"
	"html/template"
	"io"
	"path/filepath"
)

type Template string

const (
	EmailVerification Template = "email_verification.html"
	PasswordReset     Template = "password_reset.html"
	PaymentSuccess    Template = "payment_success.html"
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
func NewTemplateService(templateDir string) (TemplateService, error) {
	tmpl, err := loadTemplates(templateDir)
	if err != nil {
		return nil, err
	}
	return &templateService{templates: tmpl}, nil
}

// loadTemplates parses all templates in the specified directory.
func loadTemplates(templateDir string) (*template.Template, error) {
	files, err := filepath.Glob(filepath.Join(templateDir, "*.html"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, err
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
