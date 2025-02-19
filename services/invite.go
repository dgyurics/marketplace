package services

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/dgyurics/marketplace/repositories"
)

type contextKey string

const UserKey contextKey = "user"

// InviteService is a service for managing invitation codes.
// When REQUIRE_INVITE_CODE is set to true, users must provide a valid invitation code to register.
type InviteService interface {
	GenerateCode(ctx context.Context) (string, error)
	ValidateCode(ctx context.Context, code string, required bool) (valid bool, err error)
	StoreCode(ctx context.Context, code string, used bool) error
}

type inviteService struct {
	repo    repositories.InviteRepository
	hmacKey []byte // used for hashing invitation codes
}

func NewInviteService(repo repositories.InviteRepository, hmacKey []byte) InviteService {
	return &inviteService{
		repo:    repo,
		hmacKey: hmacKey,
	}
}

// Allowed characters for the invitation code
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

// GenerateCode creates a random 6-character alphanumeric invitation code.
func (a *inviteService) GenerateCode(ctx context.Context) (string, error) {
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

// ValidateCode retrieves an invitation code from the database and checks if it has been used.
func (a *inviteService) ValidateCode(ctx context.Context, code string, required bool) (valid bool, err error) {
	if required {
		used, _, err := a.repo.GetCode(ctx, code)
		return !used, err
	}
	return true, nil
}

// StoreCode stores an invitation code in the database. If the code already exists, it updates the "used" status.
func (a *inviteService) StoreCode(ctx context.Context, code string, used bool) error {
	return a.repo.StoreCode(ctx, hashString(code, a.hmacKey), used)
}
