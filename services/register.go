package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/url"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
	"golang.org/x/crypto/bcrypt"
)

type RegisterService interface {
	Register(ctx context.Context, email string) (string, error)
	RegisterConfirm(ctx context.Context, email, code string) error
}

type registerService struct {
	repo         repositories.RegisterRepository
	serviceEmail EmailService
	serviceTmp   TemplateService
	BaseURL      string
}

func NewRegisterService(
	repo repositories.RegisterRepository,
	serviceEmail EmailService,
	serviceTmp TemplateService,
	BaseURL string,
) RegisterService {
	return &registerService{
		repo:         repo,
		BaseURL:      BaseURL,
		serviceEmail: serviceEmail,
		serviceTmp:   serviceTmp,
	}
}

func (s *registerService) Register(ctx context.Context, email string) (string, error) {
	inUse, err := s.repo.EmailInUse(ctx, email)
	if err != nil {
		return "", err
	}
	if inUse {
		return "", types.ErrUniqueConstraintViolation
	}

	// generate registration code
	code, err := s.generateCode()
	if err != nil {
		return "", err
	}

	// hash registration code
	codeHash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// generate unique user ID
	userID, err := utilities.GenerateIDString()
	if err != nil {
		return "", err
	}

	usr := types.PendingUser{
		ID:        userID,
		Email:     email,
		CodeHash:  string(codeHash),
		ExpiresAt: time.Now().Add(3 * 24 * time.Hour),
	}
	if err := s.repo.CreatePendingUser(ctx, &usr); err != nil {
		return "", err
	}

	// Send new account registration email
	go func(email, code string) {
		detailsLink := fmt.Sprintf("%s/auth/email/%s/registration-code/%s",
			s.BaseURL,
			url.QueryEscape(email),
			url.QueryEscape(code))
		data := map[string]string{
			"DetailsLink": detailsLink,
		}
		body, err := s.serviceTmp.RenderToString(EmailVerification, data)
		if err != nil {
			slog.Error("Error loading email template: ", "error", err)
			return
		}
		payload := &types.Email{
			To:      []string{email},
			Subject: "New User Registration",
			Body:    body,
			IsHTML:  true,
		}
		if err := s.serviceEmail.Send(payload); err != nil {
			slog.Error("Error sending new user registration email: ", "email", email, "error", err)
		}
	}(email, code)

	return code, nil
}

// RegisterConfirm verifies the email and code have a matching entry in the pending_users table
func (s *registerService) RegisterConfirm(ctx context.Context, email, code string) error {
	usr, err := s.repo.GetPendingUser(ctx, email)
	if err != nil {
		return err
	}

	if usr.Used {
		slog.ErrorContext(ctx, "Registration code already used", "email", email)
		return types.ErrNotFound
	}

	if time.Now().After(usr.ExpiresAt) {
		slog.ErrorContext(ctx, "Registration code expired", "email", email)
		return types.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.CodeHash), []byte(code)); err != nil {
		slog.ErrorContext(ctx, "Invalid registration code", "email", email)
		return types.ErrNotFound
	}

	return s.repo.MarkCodeUsed(ctx, usr.ID)
}

// Allowed characters for the registration code
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

// GenerateCode creates a random 6-character alphanumeric invitation code.
func (a *registerService) generateCode() (string, error) {
	code := make([]byte, codeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", errors.New("failed to generate invite code")
		}
		code[i] = charset[num.Int64()]
	}
	return string(code), nil
}
