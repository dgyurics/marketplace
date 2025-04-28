package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/dgyurics/marketplace/repositories"
	"github.com/dgyurics/marketplace/types"
	"github.com/dgyurics/marketplace/utilities"
)

// PasswordService is the interface for password reset operations
type PasswordService interface {
	GenerateResetCode(ctx context.Context) (string, error)
	StoreResetCode(ctx context.Context, code string, email string) error
	ValidateResetCode(ctx context.Context, code, email string) (valid bool, err error)
	ResetPassword(ctx context.Context, code, email, password string) error
}

type passwordService struct {
	repo    repositories.PasswordRepository
	hmacKey []byte
}

func NewPasswordService(repo repositories.PasswordRepository, hmacKey []byte) PasswordService {
	return &passwordService{repo: repo, hmacKey: hmacKey}
}

const resetCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const resetCodeLength = 6

// GenerateResetCode generates a password reset code and returns it as a string
// The code is an alphanumeric string of length 6
func (s *passwordService) GenerateResetCode(ctx context.Context) (string, error) {
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

// StoreResetCode stores a password reset code in the database
func (s *passwordService) StoreResetCode(ctx context.Context, code string, userID string) error {
	pwdResetID, err := utilities.GenerateIDString()
	if err != nil {
		return err
	}

	return s.repo.StoreResetCode(ctx, &types.PasswordReset{
		ID:        pwdResetID,
		User:      &types.User{ID: userID},
		CodeHash:  hashString(code, s.hmacKey),
		ExpiresAt: time.Now().UTC().Add(time.Minute * 15),
	})
}

// ValidateResetCode returns true if the code is valid
func (s *passwordService) ValidateResetCode(ctx context.Context, code, email string) (valid bool, err error) {
	resetCode, err := s.repo.GetResetCode(ctx, email)
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

func (s *passwordService) ResetPassword(ctx context.Context, code, email, password string) error {
	// mark the code as used
	if err := s.repo.MarkResetCodeUsed(ctx, email); err != nil {
		return err
	}
	// update the user's password
	hashedPassword, err := generateFromPassword(password)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(ctx, email, string(hashedPassword))
}
