package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/dgyurics/marketplace/models"
	"github.com/dgyurics/marketplace/repositories"
)

type PasswordResetService interface {
	GeneratePasswordResetCode(ctx context.Context) (string, error)
	StorePasswordResetCode(ctx context.Context, code string, email string) error
	ValidatePasswordResetCode(ctx context.Context, code, email string) (valid bool, err error)
	ResetPassword(ctx context.Context, code, email, password string) error
}

type passwordResetService struct {
	repo    repositories.PasswordResetRepository
	hmacKey []byte
}

func NewPasswordResetService(repo repositories.PasswordResetRepository, hmacKey []byte) PasswordResetService {
	return &passwordResetService{repo: repo, hmacKey: hmacKey}
}

const resetCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const resetCodeLength = 6

// GeneratePasswordResetCode generates a password reset code and returns it as a string
// The code is an alphanumeric string of length 6
func (s *passwordResetService) GeneratePasswordResetCode(ctx context.Context) (string, error) {
	code := make([]byte, resetCodeLength)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(resetCodeCharset))))
		if err != nil {
			return "", errors.New("failed to generate password reset code")
		}
		code[i] = resetCodeCharset[num.Int64()]
	}
	return string(code), nil
}

// StorePasswordResetCode stores a password reset code in the database
func (s *passwordResetService) StorePasswordResetCode(ctx context.Context, code string, userID string) error {
	return s.repo.StorePasswordResetCode(ctx, &models.PasswordResetCode{
		User:      &models.User{ID: userID},
		CodeHash:  hashString(code, s.hmacKey),
		ExpiresAt: time.Now().UTC().Add(time.Minute * 15),
	})
}

// ValidatePasswordResetCode returns true if the code is valid
func (s *passwordResetService) ValidatePasswordResetCode(ctx context.Context, code, email string) (valid bool, err error) {
	resetCode, err := s.repo.GetPasswordResetCode(ctx, email)
	if err != nil {
		return false, err
	}

	// Check if the code has been used
	if resetCode.Used {
		return false, nil
	}

	// Check if the code has expired
	if time.Now().UTC().After(resetCode.ExpiresAt) {
		return false, nil
	}

	// Return true if the code matches the stored hash
	return hashString(code, s.hmacKey) == resetCode.CodeHash, nil
}

func (s *passwordResetService) ResetPassword(ctx context.Context, code, email, password string) error {
	// mark the code as used
	if err := s.repo.MarkPasswordResetCodeUsed(ctx, email); err != nil {
		return err
	}
	// update the user's password
	hashedPassword, err := generateFromPassword(password)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, email, string(hashedPassword))
}
