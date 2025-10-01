package services

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/dgyurics/marketplace/types"
)

// TODO: Refactor email architecture into two layers:
//
// 1. SMTPService (low-level) - handles raw email delivery via SMTP protocol
//    - Replaces current EmailService
//    - Responsible only for transport layer
//    - Interface: Send(email *types.Email) error
//
// 2. MailerService (high-level) - handles business email operations
//    - Aggregates SMTPService + TemplateService + config (baseURL)
//    - Provides semantic methods like SendPasswordReset(), SendOrderConfirmation()
//    - Eliminates need to inject 3 separate dependencies everywhere
//    - Single import point for all email functionality across routes/services
// type SMTPService interface {
//     Send(email *types.Email) error
// }
// type MailerService interface {
//     SendPasswordReset(recipientEmail, code string) error
//     SendPaymentSuccess(recipientEmail, orderID string) error
// }

type EmailService interface {
	Send(email *types.Email) error
}

type emailService struct {
	enabled  bool
	host     string
	port     int
	username string
	password string
	useTLS   bool
	from     string
	fromName string
}

func NewEmailService(config types.EmailConfig) EmailService {
	return &emailService{
		enabled:  config.Enabled,
		host:     config.Host,
		port:     config.Port,
		username: config.Username,
		password: config.Password,
		useTLS:   config.UseTLS,
		from:     config.From,
		fromName: config.FromName,
	}
}

func (s *emailService) Send(email *types.Email) error {
	if !s.enabled {
		slog.Warn("Email service is disabled; skipping email send", "to", email.To, "subject", email.Subject)
		return nil
	}
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))

	slog.Debug("Sending email", "to", email.To, "subject", email.Subject, "from", s.from, "host", s.host, "port", s.port, "useTLS", s.useTLS)
	if s.useTLS {
		return s.sendWithTLS(addr, email)
	}

	// Plain SMTP (for local docker containers)
	return s.sendPlain(addr, email)
}

func (s *emailService) sendPlain(addr string, email *types.Email) error {
	var a smtp.Auth
	if s.username != "" {
		a = smtp.PlainAuth("", s.username, s.password, s.host) // PORT NOT NEEDED??
	}

	msg := s.buildMessage(email)
	return smtp.SendMail(addr, a, s.from, email.To, []byte(msg))
}

func (s *emailService) sendWithTLS(addr string, email *types.Email) error {
	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS if supported
	if ok, _ := client.Extension("STARTTLS"); ok {
		slog.Debug("Starting TLS")
		config := &tls.Config{
			ServerName: s.host,
			// InsecureSkipVerify: true, // Only for testing!
		}
		if err = client.StartTLS(config); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if credentials provided
	if s.username != "" {
		slog.Debug("Authenticating", "username", s.username)
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		slog.Debug("Authentication successful")
	}

	// Set sender
	slog.Debug("Setting sender", "from", s.from)
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, addr := range email.To {
		slog.Debug("Adding recipient", "to", addr)
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	// Send message
	slog.Debug("Sending message data")
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send DATA command: %w", err)
	}

	message := s.buildMessage(email)
	if _, err = w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close message: %w", err)
	}

	slog.Debug("Email sent successfully")
	return client.Quit()
}

func (s *emailService) buildMessage(email *types.Email) string {
	var msg strings.Builder

	// Headers
	if s.fromName == "" {
		msg.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	} else {
		msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.fromName, s.from)) // WHAT'S <%s> DO??
	}

	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	// Content-Type based on whether it's HTML
	if email.IsHTML {
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	// Empty line before body
	msg.WriteString("\r\n")

	// Body
	msg.WriteString(email.Body)

	return msg.String()
}
